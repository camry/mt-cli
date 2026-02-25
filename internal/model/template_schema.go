package model

type TemplateTable struct {
    GoPackage           string
    PbPackage           string
    TableName           string
    TableNamePascalCase string
    TableNameCamelCase  string
    TableIndexList      []*TemplateTableIndex
}

// func (x *baseLevelManager) GetByLevel(level int32) (*sbpb.BaseLevel, error) {
//     k := mt.TableCacheKey(x, "GetByLevel", level)
//     if v, err := x.tableCache.Get(k); err == nil {
//         return v.(*sbpb.BaseLevel), nil
//     }
//     m := x.List()
//     for _, baseLevel := range m.GetTables() {
//         if baseLevel.GetLevel() == level {
//             err := x.tableCache.Set(k, baseLevel)
//             if err != nil {
//                 return nil, gerror.Wrapf(err, "baseLevelManager GetByLevel Set Cache Failed")
//             }
//             return baseLevel, nil
//         }
//     }
//     return nil, gerror.NewCodef(gcode.CodeNotFound, "baseLevelManager GetByLevel %v Not Found", level)
// }

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
