# dns-dependency-go
# 画出一个域名与其它DNS域的依赖关系图 
# Draw the dependency/relationship of a domain name with other DNS domains/zones
## 二进制： https://gitee.com/mchtech/dns-dependency-go/attach_files
## 效果：
### weibo.com
![image](https://gitee.com/mchtech/dns-dependency-go/raw/master/sample.min.png)
### www.amazon.com
![image](https://gitee.com/mchtech/dns-dependency-go/raw/master/complexsample.min.png)
## 用法：
> dns-dependency-go [-t] [-c] [-eip] [-nv] [-root] 域名
## 参数：
>  -t 指定DNS解析超时时间(秒，默认为2秒)

>  -c 指定DNS解析超时重试次数(默认为4)

>  -eip 指定 EDNS-Client-Subnet 的 IPv4 或 IPv6 地址

>  -nv 不验证权威记录

>  -root 解析根域名服务器记录
## 例子：

> dns-dependency-go weibo.com

> dns-dependency-go -t 2 -c 4 -eip 219.141.140.10 -root weibo.com

> dns-dependency-go -t 2 -c 4 -eip 2001:db8::1 -root weibo.com
## 依赖：
### echarts http://echarts.baidu.com/