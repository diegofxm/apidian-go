package engine

import (
	"fmt"
	"strings"
)

// ParseCreateTable genera SQL CREATE TABLE desde Operation
func ParseCreateTable(op Operation) string {
	var sql strings.Builder

	sql.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", op.Table))

	// Columnas
	for i, col := range op.Columns {
		sql.WriteString("    ")
		sql.WriteString(col.Name)
		sql.WriteString(" ")
		sql.WriteString(col.Type)

		if col.PrimaryKey {
			sql.WriteString(" PRIMARY KEY")
		}

		if col.Unique {
			sql.WriteString(" UNIQUE")
		}

		if !col.Nullable && !col.PrimaryKey {
			sql.WriteString(" NOT NULL")
		}

		if col.Default != "" {
			sql.WriteString(" DEFAULT ")
			sql.WriteString(col.Default)
		}

		if i < len(op.Columns)-1 || len(op.Constraints) > 0 || len(op.ForeignKeys) > 0 {
			sql.WriteString(",")
		}
		sql.WriteString("\n")
	}

	// Constraints
	for i, constraint := range op.Constraints {
		sql.WriteString("    ")
		sql.WriteString(fmt.Sprintf("CONSTRAINT %s ", constraint.Name))

		switch strings.ToLower(constraint.Type) {
		case "check":
			sql.WriteString(fmt.Sprintf("CHECK (%s)", constraint.Expression))
		case "unique":
			sql.WriteString(fmt.Sprintf("UNIQUE (%s)", strings.Join(constraint.Columns, ", ")))
		}

		if i < len(op.Constraints)-1 || len(op.ForeignKeys) > 0 {
			sql.WriteString(",")
		}
		sql.WriteString("\n")
	}

	// Foreign Keys
	for i, fk := range op.ForeignKeys {
		sql.WriteString("    ")
		sql.WriteString(fmt.Sprintf("CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s)",
			fk.Name, fk.Column, fk.References.Table, fk.References.Column))

		if fk.OnDelete != "" {
			sql.WriteString(fmt.Sprintf(" ON DELETE %s", fk.OnDelete))
		}
		if fk.OnUpdate != "" {
			sql.WriteString(fmt.Sprintf(" ON UPDATE %s", fk.OnUpdate))
		}

		if i < len(op.ForeignKeys)-1 {
			sql.WriteString(",")
		}
		sql.WriteString("\n")
	}

	sql.WriteString(");")

	// Comment en tabla
	if op.Comment != "" {
		sql.WriteString(fmt.Sprintf("\nCOMMENT ON TABLE %s IS '%s';", op.Table, escapeSQLString(op.Comment)))
	}

	// Indexes
	for _, idx := range op.Indexes {
		sql.WriteString("\n")
		if idx.Unique {
			sql.WriteString(fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s(%s)", idx.Name, op.Table, strings.Join(idx.Columns, ", ")))
		} else {
			sql.WriteString(fmt.Sprintf("CREATE INDEX %s ON %s(%s)", idx.Name, op.Table, strings.Join(idx.Columns, ", ")))
		}
		if idx.Where != "" {
			sql.WriteString(fmt.Sprintf(" WHERE %s", idx.Where))
		}
		sql.WriteString(";")
	}

	return sql.String()
}

// ParseDropTable genera SQL DROP TABLE
func ParseDropTable(op Operation) string {
	if op.Cascade {
		return fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE;", op.Table)
	}
	return fmt.Sprintf("DROP TABLE IF EXISTS %s;", op.Table)
}

// ParseSeed genera SQL INSERT para seeds
func ParseSeed(op Operation) []string {
	var sqls []string

	for _, row := range op.Data {
		var columns []string
		var values []string

		for col, val := range row {
			columns = append(columns, col)
			values = append(values, formatValue(val))
		}

		sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT DO NOTHING;",
			op.Table,
			strings.Join(columns, ", "),
			strings.Join(values, ", "))

		sqls = append(sqls, sql)
	}

	return sqls
}

// ParseCreateExtension genera SQL CREATE EXTENSION
func ParseCreateExtension(op Operation) string {
	return fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS \"%s\";", op.Extension)
}

// ParseCreateSequence genera SQL CREATE SEQUENCE
func ParseCreateSequence(op Operation) string {
	return fmt.Sprintf("CREATE SEQUENCE IF NOT EXISTS %s INCREMENT 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1;", op.Name)
}

// ParseDropSequence genera SQL DROP SEQUENCE
func ParseDropSequence(op Operation) string {
	if op.Cascade {
		return fmt.Sprintf("DROP SEQUENCE IF EXISTS %s CASCADE;", op.Name)
	}
	return fmt.Sprintf("DROP SEQUENCE IF EXISTS %s;", op.Name)
}

// ParseCreateTrigger genera SQL CREATE TRIGGER
func ParseCreateTrigger(op Operation) string {
	return fmt.Sprintf("CREATE TRIGGER %s %s %s ON %s FOR EACH ROW EXECUTE FUNCTION %s;",
		op.Name, op.Timing, op.Event, op.Table, op.Function)
}

// ParseDropTrigger genera SQL DROP TRIGGER
func ParseDropTrigger(op Operation) string {
	return fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s;", op.Name, op.Table)
}

// formatValue formatea un valor para SQL
func formatValue(val any) string {
	switch v := val.(type) {
	case string:
		return fmt.Sprintf("'%s'", escapeSQLString(v))
	case int, int64, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case nil:
		return "NULL"
	default:
		return fmt.Sprintf("'%v'", v)
	}
}

// escapeSQLString escapa comillas simples en strings SQL
func escapeSQLString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
