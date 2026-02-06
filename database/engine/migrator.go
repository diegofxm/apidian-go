package engine

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lib/pq"
	"gopkg.in/yaml.v3"
)

// Migrator maneja la ejecución de migraciones
type Migrator struct {
	db              *sql.DB
	tracker         *Tracker
	migrationsPath  string
	seedsPath       string
}

// NewMigrator crea un nuevo migrator
func NewMigrator(db *sql.DB, migrationsPath, seedsPath string) *Migrator {
	return &Migrator{
		db:             db,
		tracker:        NewTracker(db),
		migrationsPath: migrationsPath,
		seedsPath:      seedsPath,
	}
}

// Migrate ejecuta todas las migraciones pendientes
func (m *Migrator) Migrate() error {
	// Asegurar que existe la tabla de tracking
	if err := m.tracker.EnsureMigrationsTable(); err != nil {
		return fmt.Errorf("error creating migrations table: %w", err)
	}

	// Obtener migraciones pendientes
	pending, err := m.getPendingMigrations()
	if err != nil {
		return err
	}

	if len(pending) == 0 {
		fmt.Println("✓ No pending migrations")
		return nil
	}

	// Obtener siguiente batch
	batch, err := m.tracker.GetNextBatch()
	if err != nil {
		return err
	}

	fmt.Printf("Running %d migrations (batch %d)...\n\n", len(pending), batch)

	// Ejecutar cada migración
	for _, migFile := range pending {
		if err := m.runMigration(migFile, batch); err != nil {
			return fmt.Errorf("migration %s failed: %w", migFile, err)
		}
	}

	fmt.Println("\n✓ All migrations completed successfully")
	return nil
}

// Fresh elimina todas las tablas y ejecuta todas las migraciones
func (m *Migrator) Fresh() error {
	fmt.Println("⚠️  Dropping all tables...")

	if err := m.tracker.DropAllTables(); err != nil {
		return fmt.Errorf("error dropping tables: %w", err)
	}

	fmt.Println("✓ All tables dropped\n")

	return m.Migrate()
}

// Status muestra el estado de las migraciones
func (m *Migrator) Status() error {
	if err := m.tracker.EnsureMigrationsTable(); err != nil {
		return err
	}

	executed, err := m.tracker.GetExecutedMigrations()
	if err != nil {
		return err
	}

	allMigrations, err := m.getAllMigrationFiles()
	if err != nil {
		return err
	}

	fmt.Println("\n📊 Migration Status\n")
	fmt.Println("┌─────────────────────────────────────────────┬─────────┬───────┬─────────────────────┐")
	fmt.Println("│ Migration                                   │ Status  │ Batch │ Executed At         │")
	fmt.Println("├─────────────────────────────────────────────┼─────────┼───────┼─────────────────────┤")

	executedMap := make(map[string]MigrationRecord)
	for _, m := range executed {
		executedMap[m.Name] = m
	}

	for _, migFile := range allMigrations {
		name := strings.TrimSuffix(migFile, ".yaml")
		if record, ok := executedMap[name]; ok {
			fmt.Printf("│ %-43s │ ✓ Run   │ %5d │ %19s │\n", truncate(name, 43), record.Batch, record.ExecutedAt)
		} else {
			fmt.Printf("│ %-43s │ Pending │   -   │         -           │\n", truncate(name, 43))
		}
	}

	fmt.Println("└─────────────────────────────────────────────┴─────────┴───────┴─────────────────────┘")
	fmt.Printf("\nTotal: %d migrations (%d executed, %d pending)\n\n", len(allMigrations), len(executed), len(allMigrations)-len(executed))

	return nil
}

// Seed ejecuta todos los seeds
func (m *Migrator) Seed() error {
	seedFiles, err := m.getAllSeedFiles()
	if err != nil {
		return err
	}

	if len(seedFiles) == 0 {
		fmt.Println("✓ No seed files found")
		return nil
	}

	fmt.Printf("Running %d seed files...\n\n", len(seedFiles))

	for _, seedFile := range seedFiles {
		if err := m.runSeed(seedFile); err != nil {
			return fmt.Errorf("seed %s failed: %w", seedFile, err)
		}
	}

	// Resetear secuencias después de cargar seeds
	if err := m.resetSequences(); err != nil {
		return fmt.Errorf("error resetting sequences: %w", err)
	}

	fmt.Println("\n✓ All seeds completed successfully")
	fmt.Println("✓ Sequences reset successfully")
	return nil
}

// runMigration ejecuta una migración individual
func (m *Migrator) runMigration(filename string, batch int) error {
	name := strings.TrimSuffix(filename, ".yaml")
	fmt.Printf("→ Migrating: %s\n", name)

	// Leer archivo YAML
	migration, err := m.loadMigration(filename)
	if err != nil {
		return err
	}

	// Ejecutar operaciones UP
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, op := range migration.Up {
		sqls, err := m.operationToSQL(op)
		if err != nil {
			return err
		}

		for _, sql := range sqls {
			if _, err := tx.Exec(sql); err != nil {
				return fmt.Errorf("SQL error: %w\nQuery: %s", err, sql)
			}
		}
	}

	// Registrar migración
	if err := m.tracker.RecordMigration(name, batch); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Printf("  ✓ Migrated: %s\n", name)
	return nil
}

// runSeed ejecuta un seed individual (CSV o YAML)
func (m *Migrator) runSeed(filename string) error {
	ext := filepath.Ext(filename)

	if ext == ".csv" {
		return m.runSeedCSV(filename)
	} else if ext == ".yaml" {
		return m.runSeedYAML(filename)
	}

	return fmt.Errorf("unsupported seed format: %s", ext)
}

// runSeedYAML ejecuta un seed YAML
func (m *Migrator) runSeedYAML(filename string) error {
	name := strings.TrimSuffix(filename, ".yaml")
	fmt.Printf("→ Seeding: %s (YAML)\n", name)

	// Leer archivo YAML
	data, err := os.ReadFile(filepath.Join(m.seedsPath, filename))
	if err != nil {
		return err
	}

	var seed Seed
	if err := yaml.Unmarshal(data, &seed); err != nil {
		return err
	}

	// Crear operación de seed
	op := Operation{
		Type:  "seed",
		Table: seed.Table,
		Data:  seed.Data,
	}

	sqls := ParseSeed(op)

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, sql := range sqls {
		if _, err := tx.Exec(sql); err != nil {
			return fmt.Errorf("SQL error: %w\nQuery: %s", err, sql)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Printf("  ✓ Seeded: %s (%d rows)\n", name, len(seed.Data))
	return nil
}

// runSeedCSV ejecuta un seed CSV usando COPY FROM para máxima performance
func (m *Migrator) runSeedCSV(filename string) error {
	name := strings.TrimSuffix(filename, ".csv")
	fmt.Printf("→ Seeding: %s (CSV)\n", name)

	filePath := filepath.Join(m.seedsPath, filename)
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Leer CSV
	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("error reading CSV headers: %w", err)
	}

	// Iniciar transacción
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Usar COPY FROM para máxima performance
	stmt, err := tx.Prepare(pq.CopyIn(name, headers...))
	if err != nil {
		return fmt.Errorf("error preparing COPY: %w", err)
	}

	rowCount := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading CSV row: %w", err)
		}

		// Convertir strings a interfaces
		values := make([]interface{}, len(record))
		for i, v := range record {
			values[i] = v
		}

		if _, err := stmt.Exec(values...); err != nil {
			return fmt.Errorf("error inserting row: %w", err)
		}
		rowCount++
	}

	// Finalizar COPY
	if _, err := stmt.Exec(); err != nil {
		return fmt.Errorf("error finalizing COPY: %w", err)
	}

	if err := stmt.Close(); err != nil {
		return fmt.Errorf("error closing statement: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	fmt.Printf("  ✓ Seeded: %s (%d rows)\n", name, rowCount)
	return nil
}

// loadMigration carga una migración desde YAML
func (m *Migrator) loadMigration(filename string) (*Migration, error) {
	data, err := os.ReadFile(filepath.Join(m.migrationsPath, filename))
	if err != nil {
		return nil, err
	}

	var migration Migration
	if err := yaml.Unmarshal(data, &migration); err != nil {
		return nil, err
	}

	return &migration, nil
}

// operationToSQL convierte una operación a SQL
func (m *Migrator) operationToSQL(op Operation) ([]string, error) {
	switch op.Type {
	case "create_table":
		return []string{ParseCreateTable(op)}, nil
	case "drop_table":
		return []string{ParseDropTable(op)}, nil
	case "seed":
		return ParseSeed(op), nil
	case "raw_sql":
		return []string{op.SQL}, nil
	case "create_extension":
		return []string{ParseCreateExtension(op)}, nil
	case "create_sequence":
		return []string{ParseCreateSequence(op)}, nil
	case "drop_sequence":
		return []string{ParseDropSequence(op)}, nil
	case "create_trigger":
		return []string{ParseCreateTrigger(op)}, nil
	case "drop_trigger":
		return []string{ParseDropTrigger(op)}, nil
	default:
		return nil, fmt.Errorf("unknown operation type: %s", op.Type)
	}
}

// getPendingMigrations obtiene las migraciones pendientes
func (m *Migrator) getPendingMigrations() ([]string, error) {
	allMigrations, err := m.getAllMigrationFiles()
	if err != nil {
		return nil, err
	}

	executed, err := m.tracker.GetExecutedMigrations()
	if err != nil {
		return nil, err
	}

	executedMap := make(map[string]bool)
	for _, m := range executed {
		executedMap[m.Name] = true
	}

	var pending []string
	for _, migFile := range allMigrations {
		name := strings.TrimSuffix(migFile, ".yaml")
		if !executedMap[name] {
			pending = append(pending, migFile)
		}
	}

	return pending, nil
}

// getAllMigrationFiles obtiene todos los archivos de migración ordenados
func (m *Migrator) getAllMigrationFiles() ([]string, error) {
	files, err := os.ReadDir(m.migrationsPath)
	if err != nil {
		return nil, err
	}

	var migrations []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
			migrations = append(migrations, file.Name())
		}
	}

	sort.Strings(migrations)
	return migrations, nil
}

// getAllSeedFiles obtiene todos los archivos de seed ordenados (CSV y YAML)
func (m *Migrator) getAllSeedFiles() ([]string, error) {
	files, err := os.ReadDir(m.seedsPath)
	if err != nil {
		return nil, err
	}

	var seeds []string
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".csv")) {
			seeds = append(seeds, file.Name())
		}
	}

	sort.Strings(seeds)
	return seeds, nil
}

// resetSequences resetea todas las secuencias basándose en el MAX(id) de cada tabla
func (m *Migrator) resetSequences() error {
	// Obtener todas las secuencias del esquema public
	query := `
		SELECT 
			c.relname as sequence_name,
			REPLACE(c.relname, '_id_seq', '') as table_name
		FROM pg_class c
		WHERE c.relkind = 'S'
		AND c.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'public')
		AND c.relname LIKE '%_id_seq'
		ORDER BY c.relname
	`

	rows, err := m.db.Query(query)
	if err != nil {
		return fmt.Errorf("error querying sequences: %w", err)
	}
	defer rows.Close()

	var sequences []struct {
		sequenceName string
		tableName    string
	}

	for rows.Next() {
		var seq struct {
			sequenceName string
			tableName    string
		}
		if err := rows.Scan(&seq.sequenceName, &seq.tableName); err != nil {
			continue
		}
		sequences = append(sequences, seq)
	}

	// Resetear cada secuencia
	for _, seq := range sequences {
		// Si la tabla tiene datos, setval al MAX(id)
		// Si está vacía, resetear a 1 con is_called=false para que el próximo sea 1
		resetQuery := fmt.Sprintf(`
			DO $$
			DECLARE
				max_id BIGINT;
			BEGIN
				SELECT COALESCE(MAX(id), 0) INTO max_id FROM %s;
				IF max_id = 0 THEN
					PERFORM setval('%s', 1, false);
				ELSE
					PERFORM setval('%s', max_id, true);
				END IF;
			END $$;
		`, seq.tableName, seq.sequenceName, seq.sequenceName)
		
		if _, err := m.db.Exec(resetQuery); err != nil {
			// Ignorar si la tabla no existe
			continue
		}
	}

	return nil
}

// truncate trunca un string a una longitud máxima
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
