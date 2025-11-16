package generator

// TemplateEnum представляет структуру для шаблона
type TemplateEnum struct {
	StructType string
	Package    string
	Entries    []TemplateEntries
}

// TemplateEntries представляет поля структуры для шаблона
type TemplateEntries struct {
	Name        string
	Type        string
	IsPrimitive bool
	IsSlice     bool
	IsMap       bool
	IsStruct    bool
	IsPointer   bool
	ZeroValue   string
}
