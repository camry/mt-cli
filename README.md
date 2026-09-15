# mt-cli

表管理器代码生成工具

## 表管理器代码生成

```bash
rm -f ../mt/tables/*
mt-cli --server wild:skUbPzoif4K9uA3n@127.0.0.1:13306 --db wild_base --config ../../ola/wild-proto/sbpb/ignore.yaml --go-pkg github.com/olaola-chat/wild-pb/sbpb --out ../../ola/wild-core/mb --out-pkg mb
```
