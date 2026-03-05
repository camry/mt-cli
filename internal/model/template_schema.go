package model

type TemplateTable struct {
    GoPackage           string
    PbPackage           string
    OutPackage          string
    TableName           string
    TableNamePascalCase string
    TableNameCamelCase  string
    TableRows           int64
    TableIndexList      []*TemplateTableIndex
}

type TemplateTableIndex struct {
    IsUnique    bool
    FuncName    string
    FuncArg     string
    CacheKeyArg string
    ForIf       string
    Columns     []*TemplateTableColumn
}

type TemplateTableColumn struct {
    ColumnName           string
    PascalCaseColumnName string
    CamelCaseColumnName  string
    DataType             string
}
