<div style="text-align: center;">
  <a href="https://alist.nn.ci"><img height="100px" alt="logo" src="https://is2-ssl.mzstatic.com/image/thumb/Purple116/v4/01/45/3b/01453b82-4c0b-5140-54a3-cd60b7946f68/AppIcon-0-0-85-220-0-0-0-0-4-0-0-0-2x-sRGB-0-0-0-0-0.png/460x0w.webp"/></a>
  <p><em>Alist as a Self-hosted Private Hybrid Cloud Platform</em></p>

</div>

---

English | [中文](./README_cn.md)| [日本語](./README_ja.md) | [Contributing](./CONTRIBUTING.md) | [CODE_OF_CONDUCT](./CODE_OF_CONDUCT.md)

## Features

Alist is currently the best Cloud Storage Utilising Tool on the market in China.
To use it as a private cloud platform, privacy, security and confidentiality is a basic.
So this project should have below extra features:

- [x] Safe: a transparent encryption driver. Anyone can easily, safely store encrypted data on the remote storage provider.  Consider your data is safely stored in the safe, and storage provider can only see the safe, but not your data.
  - [x] Optional: compatible with [Rclone Crypt](https://rclone.org/crypt/). More ways to manipulate the encrypted data.
  - [x] directory and filename encryption
- [x] obfuscate sensitive information in config files
- [x] server-side encryption mode (server encrypt & decrypt all data, all data flows thru server)
- [ ] client side encryption mode (client browser request data from remote storage cloud then decrypt data in client)
  - [ ] IOS/Mac: Quantumult X/Surge/ShadowRocket Box.js script. Hijack access to all Cloud Storage, also alist server will always return 302 redirect, so when app accessing alist, redirecting to storage Provider, and the header is changed by the script. any app will work with the webdav.
  - [ ] Android/Mac/Windows/Linux/Docker : Mini local proxy for Alist server, so any app/client with in the lan/device will directly access
- [x] multi-thread downloading for [Quark] Drive, will add option of enable for other slow drivers. (high-memory usage but better experience in non-multithreading download tools, e.g. playing media in any media player)
- [ ] auto retry in background Task: Move, Copy operations.
  - Task resume from last failed point
  - Task auto retry on failure
  - Task size validates on finish
- [ ] rapid-upload support move between local/smb and cloud storage with SHA1 hash return. 

Other features please refer to [Alist](https://github.com/alist-org/alist)

## Client Development Plan

| Client Platform               | Speed*                  | Dependency                                 | Compatibility                     | Comment                     | Available?                                                                                        |
|-------------------------------|-------------------------|--------------------------------------------|-----------------------------------|-----------------------------|---------------------------------------------------------------------------------------------------|
| IOS/Mac                       | Max                     | Quantumult X/Surge/ShadowRocket [Paid App] | Browser, Any app work with webdav | data hijack for some url    | work in progress*                                                                                 |
| Android                       | Max                     | Client App running in the background       | Browser, Any app work with webdav | client proxy data           |                                                                                                   |
| Mac/Windows/Linux/Docker      | Max                     | Application running in the background      | Browser, Any app work with webdav | client proxy data           |                                                                                                   |
| Web access                    | Limited by server       |                                            | Browser,Any app work with webdav  |                             | Yes                                                                                               |
| Web access[Alist Hybrid mode] | Max                     |                                            | Browser                           | use JS to make http request |                                                                                                   |
| Web access with alist-proxy   | Limited by proxy server | server with high bandwidth, domain         | Browser, Any app work with webdav | another server proxy data   | Yes with no Vault support                                                                         |
| Native IOS/Android App        | Limited by server       |                                            | app internal function             | app has limited function    | [xlist](https://github.com/xlist-io/xlist) [AlistClient](https://github.com/BFWXKJGS/AlistClient) |

* 1 Assume client & cloud storage provider has unlimited bandwidth.
* 2 Details on the one I'm working on: after configured the rule in those apps, when accessing the alist server, the rule will get the 302 redirect url ,then apply the necessary header, and request the url from the device.


| 客户端平台                    | 速度*      | 依赖性                                    | 兼容性                | 其他               | 可用的？                                                                                                |
|--------------------------|----------|----------------------------------------|--------------------|------------------|-----------------------------------------------------------------------------------------------------|
| IOS/Mac                  | 最大       | Quantumult X/Surge/Shadowrocket [付费应用] | 浏览器，WebDav兼容的任何app | 特定url数据劫持        | 正在开发*                                                                                               |
| Android                  | 最大       | 客户端后台运行                                | 浏览器，WebDav兼容的任何app | 客户端代理数据          |                                                                                                     |
| Mac/Windows/Linux/Docker | 最大       | 后台运行的应用程序                              | 浏览器，WebDav兼容的任何app | 客户端代理数据          |                                                                                                     |
| Web访问                    | 受服务器限制   |                                        | 浏览器，WebDav兼容的任何app |                  | 是                                                                                                   |
| Web访问[Alist Hybrid模式]    | 最大       |                                        | 浏览器                | 使用浏览器脚本js来下载远程数据 |                                                                                                     |
| Web访问Alist-Proxy         | 受代理服务器限制 | 大带宽服务器，域名                              | 浏览器，WebDav兼容的任何app | 另一个服务器代理数据       | 是，但没有加密支持                                                                                           |
| 原生IOS/Mac 应用             | 受服务器限制   |                                        | 应用程序内部功能           | 应用功能有限           | [xlist](https://github.com/xlist-io/xlist)   [AlistClient](https://github.com/BFWXKJGS/AlistClient) |

* 1 假设客户端和云存储提供商具有无限的带宽。
* 2 我正在开发的流量劫持规则：在这些vpn app中配置规则后，在访问ALIST服务器时，该规则将获得302重定向URL，然后修改数据包(header)后直接从设备请求URL。

## License

The `AList-Private-Cloud` is open-source software licensed under the AGPL-3.0 license.

## Last but not least
- This fork is a free and open source project.
-

---