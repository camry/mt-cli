package model

type TemplateTable struct {
    GoPackage             string
    PbPackage             string
    OutPackage            string
    TableName             string
    TableNamePascalCase   string
    TableNameCamelCase    string
    TableRows             int64
    TableIndexList        []*TemplateTableIndex
    TableIndexFuncNameLen int
}

type TemplateTableIndex struct {
    IsUnique          bool
    FuncName          string
    FuncNameCamelCase string
    FuncArg           string
    FuncArgName       string
    FuncArgLen        int
    FuncKeyF          string
    FuncKeyArg        string
    ForIf             string
    Columns           []*TemplateTableColumn
    ColumnLenMax      int
}

type TemplateTableColumn struct {
    ColumnName           string
    PascalCaseColumnName string
    CamelCaseColumnName  string
    DataType             string
}
