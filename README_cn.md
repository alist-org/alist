<div align="center">
  <a href="https://alist.nn.ci"><img height="100px" alt="logo" src="https://cdn.jsdelivr.net/gh/alist-org/logo@main/logo.svg"/></a>
  <p><em>🗂一个支持多存储的文件列表程序，使用 Gin 和 Solidjs。</em></p>
<div>
  <a href="https://goreportcard.com/report/github.com/alist-org/alist/v3">
    <img src="https://goreportcard.com/badge/github.com/alist-org/alist/v3" alt="latest version" />
  </a>
  <a href="https://github.com/alist-org/alist/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/Xhofe/alist" alt="License" />
  </a>
  <a href="https://github.com/alist-org/alist/actions?query=workflow%3ABuild">
    <img src="https://img.shields.io/github/actions/workflow/status/Xhofe/alist/build.yml?branch=main" alt="Build status" />
  </a>
  <a href="https://github.com/alist-org/alist/releases">
    <img src="https://img.shields.io/github/release/Xhofe/alist" alt="latest version" />
  </a>
  <a title="Crowdin" target="_blank" href="https://crwd.in/alist">
    <img src="https://badges.crowdin.net/alist/localized.svg">
  </a>
</div>
<div>
  <a href="https://github.com/alist-org/alist/discussions">
    <img src="https://img.shields.io/github/discussions/Xhofe/alist?color=%23ED8936" alt="discussions" />
  </a>
  <a href="https://discord.gg/F4ymsH4xv2">
    <img src="https://img.shields.io/discord/1018870125102895134?logo=discord" alt="discussions" />
  </a>
  <a href="https://github.com/alist-org/alist/releases">
    <img src="https://img.shields.io/github/downloads/Xhofe/alist/total?color=%239F7AEA&logo=github" alt="Downloads" />
  </a>
  <a href="https://hub.docker.com/r/xhofe/alist">
    <img src="https://img.shields.io/docker/pulls/xhofe/alist?color=%2348BB78&logo=docker&label=pulls" alt="Downloads" />
  </a>
  <a href="https://alist.nn.ci/zh/guide/sponsor.html">
    <img src="https://img.shields.io/badge/%24-sponsor-F87171.svg" alt="sponsor" />
  </a>
</div>
</div>

---

[English](./README.md) | 中文 | [日本語](./README_ja.md) | [Contributing](./CONTRIBUTING.md) | [CODE_OF_CONDUCT](./CODE_OF_CONDUCT.md)

## 功能

- [x] 多种存储
    - [x] 本地存储
    - [x] [阿里云盘](https://www.alipan.com/)
    - [x] OneDrive / Sharepoint（[国际版](https://www.office.com/), [世纪互联](https://portal.partner.microsoftonline.cn),de,us）
    - [x] [天翼云盘](https://cloud.189.cn) (个人云, 家庭云)
    - [x] [GoogleDrive](https://drive.google.com/)
    - [x] [123云盘](https://www.123pan.com/)
    - [x] FTP / SFTP
    - [x] [PikPak](https://www.mypikpak.com/)
    - [x] [S3](https://aws.amazon.com/cn/s3/)
    - [x] [Seafile](https://seafile.com/)
    - [x] [又拍云对象存储](https://www.upyun.com/products/file-storage)
    - [x] WebDav(支持无API的OneDrive/SharePoint)
    - [x] Teambition（[中国](https://www.teambition.com/ )，[国际](https://us.teambition.com/ )）
    - [x] [分秒帧](https://www.mediatrack.cn/)
    - [x] [和彩云](https://yun.139.com/) (个人云, 家庭云)
    - [x] [Yandex.Disk](https://disk.yandex.com/)
    - [x] [百度网盘](http://pan.baidu.com/)
    - [x] [UC网盘](https://drive.uc.cn)
    - [x] [夸克网盘](https://pan.quark.cn)
    - [x] [迅雷网盘](https://pan.xunlei.com)
    - [x] [蓝奏云](https://www.lanzou.com/)
    - [x] [蓝奏云优享版](https://www.ilanzou.com/)
    - [x] [阿里云盘分享](https://www.alipan.com/)
    - [x] [谷歌相册](https://photos.google.com/)
    - [x] [Mega.nz](https://mega.nz)
    - [x] [一刻相册](https://photo.baidu.com/)
    - [x] SMB
    - [x] [115](https://115.com/)
    - [X] Cloudreve
    - [x] [Dropbox](https://www.dropbox.com/)
    - [x] [飞机盘](https://www.feijipan.com/)
    - [x] [多吉云](https://www.dogecloud.com/product/oss)
- [x] 部署方便，开箱即用
- [x] 文件预览（PDF、markdown、代码、纯文本……）
- [x] 画廊模式下的图像预览
- [x] 视频和音频预览，支持歌词和字幕
- [x] Office 文档预览（docx、pptx、xlsx、...）
- [x] `README.md` 预览渲染
- [x] 文件永久链接复制和直接文件下载
- [x] 黑暗模式
- [x] 国际化
- [x] 受保护的路由（密码保护和身份验证）
- [x] WebDav (具体见 https://alist.nn.ci/zh/guide/webdav.html)
- [x] [Docker 部署](https://hub.docker.com/r/xhofe/alist)
- [x] Cloudflare workers 中转
- [x] 文件/文件夹打包下载
- [x] 网页上传(可以允许访客上传)，删除，新建文件夹，重命名，移动，复制
- [x] 离线下载
- [x] 跨存储复制文件
- [x] 单线程下载/串流的多线程下载加速

## 文档

<https://alist.nn.ci/zh/>

## Demo

<https://al.nn.ci>

## 讨论

一般问题请到[讨论论坛](https://github.com/alist-org/alist/discussions) ，**issue仅针对错误报告和功能请求。**

## 赞助

AList 是一个开源软件，如果你碰巧喜欢这个项目，并希望我继续下去，请考虑赞助我或提供一个单一的捐款！感谢所有的爱和支持：https://alist.nn.ci/zh/guide/sponsor.html

### 特别赞助

- [VidHub](https://zh.okaapps.com/product/1659622164?ref=alist) - 苹果生态下优雅的网盘视频播放器，iPhone，iPad，Mac，Apple TV全平台支持。
- [亚洲云](https://www.asiayun.com/aff/QQCOOQKZ) - 高防服务器|服务器租用|福州高防|广东电信|香港服务器|美国服务器|海外服务器 - 国内靠谱的企业级云计算服务提供商 (国内API服务器赞助)
- [找资源](https://zhaoziyuan.pw/) - 阿里云盘资源搜索引擎

## 贡献者

Thanks goes to these wonderful people:

[![Contributors](http://contrib.nn.ci/api?repo=alist-org/alist&repo=alist-org/alist-web&repo=alist-org/docs)](https://github.com/alist-org/alist/graphs/contributors)

## 许可

`AList` 是在 AGPL-3.0 许可下许可的开源软件。

## 免责声明
- 本程序为免费开源项目，旨在分享网盘文件，方便下载以及学习golang，使用时请遵守相关法律法规，请勿滥用；
- 本程序通过调用官方sdk/接口实现，无破坏官方接口行为；
- 本程序仅做302重定向/流量转发，不拦截、存储、篡改任何用户数据；
- 在使用本程序之前，你应了解并承担相应的风险，包括但不限于账号被ban，下载限速等，与本程序无关；
- 如有侵权，请通过[邮件](mailto:i@nn.ci)与我联系，会及时处理。

---

> [@博客](https://nn.ci/) · [@GitHub](https://github.com/alist-org) · [@Telegram群](https://t.me/alist_chat) · [@Discord](https://discord.gg/F4ymsH4xv2)
