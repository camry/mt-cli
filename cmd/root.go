package cmd

import (
    _ "embed"
    "os"
    "strings"
    "sync"
    "text/template"

    "github.com/camry/g/v2/gerrors/gerror"
    "github.com/camry/g/v2/glog"
    "github.com/samber/lo"
    "github.com/spf13/cobra"
    "gopkg.in/yaml.v3"

    "gitea.meigushi.go/meigushi/mt-cli/internal/gdb"
    "gitea.meigushi.go/meigushi/mt-cli/internal/gen"
)

var (
    //go:embed .version
    Version string
    //go:embed template/tables.tpl
    TemplateTablesTpl string
)

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    rootCmd.Flags().StringVarP(&server, "server", "s", "", "指定服务器。(格式: <user>:<password>@<host>:<port>)")
    rootCmd.Flags().StringVarP(&db, "db", "d", "", "指定数据库。")
    rootCmd.Flags().StringVarP(&cfgPath, "config", "c", "", "指定配置文件路径。")
    rootCmd.Flags().StringVarP(&out, "out", "o", "", "指定输出目录。")
    rootCmd.Flags().StringVarP(&goPkg, "go-pkg", "g", "", "指定 Go 包名。")
    rootCmd.Flags().StringVarP(&outPkg, "out-pkg", "p", "tables", "指定输出 Go 包名。")

    cobra.CheckErr(rootCmd.MarkFlagRequired("server"))
    cobra.CheckErr(rootCmd.MarkFlagRequired("db"))
    cobra.CheckErr(rootCmd.MarkFlagRequired("out"))
    cobra.CheckErr(rootCmd.MarkFlagRequired("go-pkg"))
}

type Config struct {
    Ignores []*IgnoreTable `yaml:"ignores"`
}

type IgnoreTable struct {
    Table   string   `yaml:"table"`
    Columns []string `yaml:"columns"`
}

var (
    server  string
    db      string
    cfgPath string
    out     string
    goPkg   string
    outPkg  string

    ch chan struct{}
    wg sync.WaitGroup

    rootCmd = &cobra.Command{
        Use:     "mt-cli",
        Short:   "表管理器代码生成工具。",
        Version: Version,
        Run: func(cmd *cobra.Command, args []string) {
            mtDB, err := gdb.NewMtDB(server, db)
            if err != nil {
                glog.Fatal(gerror.Wrapf(err, "NewMtDB Failed"))
            }

            pbPkg, ok := lo.Last(strings.Split(goPkg, "/"))
            if !ok {
                glog.Fatal(gerror.Newf("--go-pkg 参数错误（例：github.com/camry/pb/sbpb）。"))
            }
            tpl, err := template.New("TemplateTablesTpl").Parse(TemplateTablesTpl)
            if err != nil {
                glog.Fatal(gerror.Wrapf(err, "template.New Failed"))
            }

            // Load Ignore Config
            icMap := make(map[string]*IgnoreTable, 50)
            icMap["base_json"] = &IgnoreTable{
                Table: "base_json",
            }
            icMap["base_json_proto"] = &IgnoreTable{
                Table: "base_json_proto",
            }
            icMap["base_json_proto_column"] = &IgnoreTable{
                Table: "base_json_proto_column",
            }
            if cfgPath != "" {
                var ic *Config
                bytes, err := os.ReadFile(cfgPath)
                if err != nil {
                    glog.Fatal(gerror.Wrapf(err, "ReadFile Failed"))
                }
                err = yaml.Unmarshal(bytes, &ic)
                if err != nil {
                    glog.Fatal(gerror.Wrapf(err, "yaml.Unmarshal Failed"))
                }

                for _, vv := range ic.Ignores {
                    if len(vv.Columns) > 0 {
                        continue
                    }
                    icMap[vv.Table] = vv
                }
            }

            ch = make(chan struct{}, 10)
            for _, table := range mtDB.GetTables() {
                ok1 := !strings.HasPrefix(table.TableName, "base_")
                _, ok2 := icMap[table.TableName]
                if ok1 || ok2 {
                    continue
                }
                ch <- struct{}{}
                g := gen.NewTable(mtDB, tpl, table, goPkg, pbPkg, outPkg, out)
                wg.Go(func() {
                    if err := g.Generate(); err != nil {
                        glog.Fatal(gerror.Wrapf(err, "Generate Failed"))
                    }
                    <-ch
                })
            }
            wg.Wait()
        },
    }
)
