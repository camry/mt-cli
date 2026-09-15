package gen

import (
    "bytes"
    "fmt"
    "os"
    "sort"
    "strings"
    "text/template"

    "github.com/camry/g/v2/gerrors/gerror"
    "github.com/samber/lo"

    "gitea.meigushi.go/meigushi/mt-cli/internal/gdb"
    "gitea.meigushi.go/meigushi/mt-cli/internal/model"
)

type Table struct {
    mtDB       *gdb.MtDB
    tpl        *template.Template
    modelTable *model.Table
    goPkg      string
    pbPkg      string
    outPkg     string
    out        string
}

func NewTable(mtDB *gdb.MtDB, tpl *template.Template, modelTable *model.Table, goPkg, pbPkg, outPkg, out string) *Table {
    return &Table{
        mtDB:       mtDB,
        tpl:        tpl,
        modelTable: modelTable,
        goPkg:      goPkg,
        pbPkg:      pbPkg,
        outPkg:     outPkg,
        out:        out,
    }
}

func (x *Table) Generate() error {
    templateTable := &model.TemplateTable{
        GoPackage:           x.goPkg,
        PbPackage:           x.pbPkg,
        OutPackage:          x.outPkg,
        TableName:           x.modelTable.TableName,
        TableNamePascalCase: lo.PascalCase(x.modelTable.TableName),
        TableNameCamelCase:  lo.CamelCase(x.modelTable.TableName),
    }

    var (
        tableRows          int64
        tableColumnList    []*model.Column
        tableStatisticList []*model.Statistic
    )

    if x.modelTable.TableRows.Valid {
        tableRows = x.modelTable.TableRows.Int64
    }

    err := x.mtDB.GetDB().Table("information_schema.COLUMNS").Order("`ORDINAL_POSITION` ASC").Find(
        &tableColumnList,
        "`TABLE_SCHEMA` = ? AND `TABLE_NAME` = ?", x.mtDB.GetSchema().SchemaName, x.modelTable.TableName,
    ).Error
    if err != nil {
        return gerror.Wrapf(err, "Table Generate Find COLUMNS By TABLE_SCHEMA=%s,TABLE_NAME=%s Failed", x.mtDB.GetSchema().SchemaName, x.modelTable.TableName)
    }
    err = x.mtDB.GetDB().Table("information_schema.STATISTICS").Order("`INDEX_NAME`,`SEQ_IN_INDEX` ASC").Find(
        &tableStatisticList,
        "`TABLE_SCHEMA` = ? AND `TABLE_NAME` = ?", x.mtDB.GetSchema().SchemaName, x.modelTable.TableName,
    ).Error
    if err != nil {
        return gerror.Wrapf(err, "Table Generate Find STATISTICS By TABLE_SCHEMA=%s,TABLE_NAME=%s Failed", x.mtDB.GetSchema().SchemaName, x.modelTable.TableName)
    }

    if len(tableColumnList) == 0 || len(tableStatisticList) == 0 {
        return gerror.Newf("Table Generate %s Not Found", x.modelTable.TableName)
    }
    tableColumnMap := lo.SliceToMap(tableColumnList, func(item *model.Column) (string, *model.Column) {
        item.DataType = strings.ToUpper(item.DataType)
        item.ColumnType = strings.ToUpper(item.ColumnType)
        return item.ColumnName, item
    })
    tableStatisticGroup := lo.GroupBy(tableStatisticList, func(item *model.Statistic) string {
        return item.IndexName
    })
    tableStatisticGroupIndexNames := lo.Keys(tableStatisticGroup)
    sort.Strings(tableStatisticGroupIndexNames)
    templateTable.TableIndexList = make([]*model.TemplateTableIndex, 0, len(tableStatisticGroup))
    for _, tableStatisticGroupIndexName := range tableStatisticGroupIndexNames {
        indexTableStatisticList := tableStatisticGroup[tableStatisticGroupIndexName]
        templateTableIndex := &model.TemplateTableIndex{
            Columns: make([]*model.TemplateTableColumn, 0, len(indexTableStatisticList)),
        }
        isUnique := false
        pascalCaseColumnNames := make([]string, 0, len(indexTableStatisticList))
        camelCaseColumnNames := make([]string, 0, len(indexTableStatisticList))

        for _, indexTableStatistic := range indexTableStatisticList {
            tableColumn, ok := tableColumnMap[indexTableStatistic.ColumnName]
            if !ok {
                return gerror.Newf("Table Generate %s ColumnName %s Not Found", x.modelTable.TableName, indexTableStatistic.ColumnName)
            }
            templateTableColumn := &model.TemplateTableColumn{
                ColumnName:           indexTableStatistic.ColumnName,
                PascalCaseColumnName: lo.PascalCase(indexTableStatistic.ColumnName),
                CamelCaseColumnName:  lo.CamelCase(indexTableStatistic.ColumnName),
                DataType:             x.DataType(tableColumn),
            }
            templateTableIndex.Columns = append(templateTableIndex.Columns, templateTableColumn)
            pascalCaseColumnNames = append(pascalCaseColumnNames, templateTableColumn.PascalCaseColumnName)
            camelCaseColumnNames = append(camelCaseColumnNames, templateTableColumn.CamelCaseColumnName)
            isUnique = indexTableStatistic.NonUnique == 0
        }
        templateTableIndex.FuncName = strings.Join(pascalCaseColumnNames, "And")
        if isUnique {
            templateTableIndex.FuncName = "GetBy" + templateTableIndex.FuncName
        } else {
            templateTableIndex.FuncName = "ListBy" + templateTableIndex.FuncName
        }
        templateTableIndex.FuncNameCamelCase = lo.CamelCase(templateTableIndex.FuncName)
        templateTable.TableIndexFuncNameLen = lo.Max([]int{templateTable.TableIndexFuncNameLen, len(templateTableIndex.FuncName)})
        templateTableIndex.IsUnique = isUnique
        columnArgs := make([]string, 0, len(templateTableIndex.Columns))
        columnArgNames := make([]string, 0, len(templateTableIndex.Columns))
        columnKeyFs := make([]string, 0, len(templateTableIndex.Columns))
        columnKeyArgs := make([]string, 0, len(templateTableIndex.Columns))
        ifColumns := make([]string, 0, len(templateTableIndex.Columns))
        for _, tableColumn := range templateTableIndex.Columns {
            templateTableIndex.ColumnLenMax = lo.Max([]int{templateTableIndex.ColumnLenMax, len(tableColumn.CamelCaseColumnName)})
            columnArgs = append(columnArgs, fmt.Sprintf("%s %s", tableColumn.CamelCaseColumnName, tableColumn.DataType))
            columnArgNames = append(columnArgNames, tableColumn.CamelCaseColumnName)

            switch tableColumn.DataType {
            case "int32", "int64":
                columnKeyFs = append(columnKeyFs, tableColumn.PascalCaseColumnName+"=%d")
            case "float", "double":
                columnKeyFs = append(columnKeyFs, tableColumn.PascalCaseColumnName+"=%.4f")
            case "bool":
                columnKeyFs = append(columnKeyFs, tableColumn.PascalCaseColumnName+"=%t")
            case "string":
                columnKeyFs = append(columnKeyFs, tableColumn.PascalCaseColumnName+"=%s")
            default:
                columnKeyFs = append(columnKeyFs, tableColumn.PascalCaseColumnName+"=%v")
            }

            columnKeyArgs = append(columnKeyArgs, tableColumn.CamelCaseColumnName)
            ifColumns = append(ifColumns, fmt.Sprintf("%s.Get%s() == %s", templateTable.TableNameCamelCase, tableColumn.PascalCaseColumnName, tableColumn.CamelCaseColumnName))
        }
        templateTableIndex.FuncArg = strings.Join(columnArgs, ", ")
        templateTableIndex.FuncArgName = strings.Join(columnArgNames, ", ")
        templateTableIndex.FuncArgLen = len(templateTableIndex.Columns)
        templateTableIndex.FuncKeyF = strings.Join(columnKeyFs, ", ")
        templateTableIndex.FuncKeyArg = strings.Join(columnKeyArgs, ", ")
        templateTableIndex.ForIf = strings.Join(ifColumns, " && ")
        templateTable.TableIndexList = append(templateTable.TableIndexList, templateTableIndex)
    }
    templateTable.TableRows = tableRows

    var wr bytes.Buffer
    err = x.tpl.Execute(&wr, templateTable)
    if err != nil {
        return gerror.Wrapf(err, "tpl.Execute Failed")
    }
    err = os.WriteFile(fmt.Sprintf("%s/%s.go", x.out, x.modelTable.TableName), wr.Bytes(), 0644)
    if err != nil {
        return gerror.Wrapf(err, "os.WriteFile Failed")
    }
    return nil
}

func (x *Table) DataType(tableColumn *model.Column) string {
    switch tableColumn.DataType {
    case "TINYINT", "SMALLINT", "MEDIUMINT", "INT", "INTEGER":
        if tableColumn.ColumnType == "TINYINT(1)" {
            return "bool"
        }
        return "int32"
    case "BIGINT":
        return "int64"
    case "FLOAT", "DECIMAL":
        return "float"
    case "DOUBLE":
        return "double"
    case "TINYBLOB", "BLOB", "MEDIUMBLOB", "LONGBLOB":
        return "bytes"
    default:
        return "string"
    }
}
