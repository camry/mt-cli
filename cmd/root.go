package cmd

import (
    _ "embed"
    "strings"
    "sync"
    "text/template"

    "github.com/camry/g/v2/gerrors/gerror"
    "github.com/camry/g/v2/glog"
    "github.com/samber/lo"
    "github.com/spf13/cobra"

    "mt-cli/internal/gdb"
    "mt-cli/internal/gen"
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
    rootCmd.Flags().StringVarP(&out, "out", "o", "", "指定输出目录。")
    rootCmd.Flags().StringVarP(&goPkg, "go-pkg", "g", "", "指定 Go 包名。")

    cobra.CheckErr(rootCmd.MarkFlagRequired("server"))
    cobra.CheckErr(rootCmd.MarkFlagRequired("db"))
    cobra.CheckErr(rootCmd.MarkFlagRequired("out"))
    cobra.CheckErr(rootCmd.MarkFlagRequired("go-pkg"))
}

var (
    server string
    db     string
    out    string
    goPkg  string

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

            ch = make(chan struct{}, 10)
            for _, table := range mtDB.GetTables() {
                c1 := !strings.HasPrefix(table.TableName, "base_")
                c2 := strings.HasPrefix(table.TableName, "base_json")
                if c1 || c2 {
                    continue
                }
                ch <- struct{}{}
                g := gen.NewTable(mtDB, tpl, table, goPkg, pbPkg, out)
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
