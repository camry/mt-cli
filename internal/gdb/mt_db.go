package gdb

import (
    "fmt"
    "regexp"
    "strconv"
    "strings"

    "github.com/camry/g/v2/gerrors/gerror"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"

    "mt-cli/internal/def"
    "mt-cli/internal/model"
)

type MtDB struct {
    db     *gorm.DB
    schema *model.Schema
    tables []*model.Table
}

func NewMtDB(server string, database string) (*MtDB, error) {
    serverMatched, err := regexp.MatchString(def.HostPattern, server)
    if err != nil {
        return nil, err
    }
    dbMatched, err := regexp.MatchString(def.DbPattern, database)
    if err != nil {
        return nil, err
    }
    if !serverMatched {
        return nil, gerror.Newf("服务器 `%s` 格式错误。(正确格式: <user>:<password>@<host>:<port>)", server)
    }
    if !dbMatched {
        return nil, gerror.Newf("数据库 `%s` 格式错误。", database)
    }

    var (
        serverUser = strings.Split(server[0:strings.LastIndex(server, "@")], ":")
        serverHost = strings.Split(server[strings.LastIndex(server, "@")+1:], ":")
    )
    serverDbConfig := &model.DbConfig{
        User:     serverUser[0],
        Password: serverUser[1],
        Host:     serverHost[0],
        Charset:  "utf8mb4",
        Database: database,
    }
    serverDbConfig.Port, err = strconv.Atoi(serverHost[1])
    if err != nil {
        return nil, err
    }

    serverDb, err := gorm.Open(mysql.New(mysql.Config{
        DSN: fmt.Sprintf(def.Dsn,
            serverDbConfig.User, serverDbConfig.Password,
            serverDbConfig.Host, serverDbConfig.Port,
            serverDbConfig.Database, serverDbConfig.Charset,
        ),
    }), &gorm.Config{
        SkipDefaultTransaction: true,
        DisableAutomaticPing:   true,
        Logger:                 logger.Default.LogMode(logger.Silent),
    })
    if err != nil {
        return nil, err
    }

    serverSchema := &model.Schema{}
    serverSchemaResult := serverDb.Table("information_schema.SCHEMATA").Limit(1).Find(
        serverSchema,
        "`SCHEMA_NAME` = ?", serverDbConfig.Database,
    )
    if serverSchemaResult.RowsAffected <= 0 {
        return nil, gerror.Newf("数据库 `%s` 不存在。", serverDbConfig.Database)
    }

    var (
        tables []*model.Table
    )
    serverTableResult := serverDb.Table("information_schema.TABLES").Order("`TABLE_NAME` ASC").Find(
        &tables,
        "`TABLE_SCHEMA` = ?", serverDbConfig.Database,
    )
    if serverTableResult.RowsAffected <= 0 {
        return nil, gerror.Newf("数据库 `%s` 没有表。", serverDbConfig.Database)
    }

    return &MtDB{
        db:     serverDb,
        schema: serverSchema,
        tables: tables,
    }, nil
}

func (x *MtDB) GetDB() *gorm.DB {
    return x.db
}

func (x *MtDB) GetSchema() *model.Schema {
    return x.schema
}

func (x *MtDB) GetTables() []*model.Table {
    return x.tables
}
