# xwc-controller


provider    -- precher    -- precheck

            -- installer  -- install    -- to call controller-agent module
                          -- reset
                          -- scale
                          -- reduce
                          
## 启动

### 基于配置文件创建configMap
名称为pwc-controller-config

### 创建yaml文件
通过挂载configMap将配置文件挂载到容器内

通过指定容器参数设置配置文件
- -config
- /etc/config/config.json  

