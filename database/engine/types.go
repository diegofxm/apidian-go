package engine

// Migration representa una migración completa desde YAML
type Migration struct {
	Version     string      `yaml:"version"`
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Up          []Operation `yaml:"up"`
	Down        []Operation `yaml:"down"`
}

// Operation representa una operación SQL (crear tabla, insertar datos, etc.)
type Operation struct {
	Type        string            `yaml:"type"` // create_table, drop_table, seed, raw_sql, create_extension, create_sequence, drop_sequence, create_trigger, drop_trigger
	Table       string            `yaml:"table,omitempty"`
	Name        string            `yaml:"name,omitempty"`
	Extension   string            `yaml:"extension,omitempty"`
	Columns     []Column          `yaml:"columns,omitempty"`
	Constraints []Constraint      `yaml:"constraints,omitempty"`
	Indexes     []Index           `yaml:"indexes,omitempty"`
	ForeignKeys []ForeignKey      `yaml:"foreign_keys,omitempty"`
	Comment     string            `yaml:"comment,omitempty"`
	Data        []map[string]any  `yaml:"data,omitempty"`
	SQL         string            `yaml:"sql,omitempty"`
	Cascade     bool              `yaml:"cascade,omitempty"`
	Timing      string            `yaml:"timing,omitempty"`   // BEFORE, AFTER
	Event       string            `yaml:"event,omitempty"`    // INSERT, UPDATE, DELETE
	Function    string            `yaml:"function,omitempty"` // Nombre de función para trigger
}

// Column representa una columna de tabla
type Column struct {
	Name       string `yaml:"name"`
	Type       string `yaml:"type"`
	Nullable   bool   `yaml:"nullable,omitempty"`
	PrimaryKey bool   `yaml:"primary_key,omitempty"`
	Unique     bool   `yaml:"unique,omitempty"`
	Default    string `yaml:"default,omitempty"`
	Comment    string `yaml:"comment,omitempty"`
}

// Constraint representa un constraint (CHECK, UNIQUE, etc.)
type Constraint struct {
	Type       string   `yaml:"type"` // check, unique
	Name       string   `yaml:"name"`
	Expression string   `yaml:"expression,omitempty"` // Para CHECK
	Columns    []string `yaml:"columns,omitempty"`    // Para UNIQUE
}

// Index representa un índice
type Index struct {
	Name    string   `yaml:"name"`
	Columns []string `yaml:"columns"`
	Unique  bool     `yaml:"unique,omitempty"`
	Where   string   `yaml:"where,omitempty"`
}

// ForeignKey representa una clave foránea
type ForeignKey struct {
	Name       string    `yaml:"name"`
	Column     string    `yaml:"column"`
	References Reference `yaml:"references"`
	OnDelete   string    `yaml:"on_delete,omitempty"` // CASCADE, RESTRICT, SET NULL
	OnUpdate   string    `yaml:"on_update,omitempty"`
}

// Reference representa la referencia de una FK
type Reference struct {
	Table  string `yaml:"table"`
	Column string `yaml:"column"`
}

// Seed representa datos de seed desde YAML
type Seed struct {
	Table string              `yaml:"table"`
	Data  []map[string]any    `yaml:"data"`
}

// MigrationRecord representa un registro en schema_migrations
type MigrationRecord struct {
	ID        int
	Name      string
	Batch     int
	ExecutedAt string
}
