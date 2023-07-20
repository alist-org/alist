<div style="text-align: center;">
  <a href="https://alist.nn.ci"><img height="100px" alt="logo" src="https://is2-ssl.mzstatic.com/image/thumb/Purple116/v4/01/45/3b/01453b82-4c0b-5140-54a3-cd60b7946f68/AppIcon-0-0-85-220-0-0-0-0-4-0-0-0-2x-sRGB-0-0-0-0-0.png/460x0w.webp"/></a>
  <p><em>Alist as a Self-hosted Private Hybrid Cloud Platform</em></p>

</div>

---

## Features

Alist is currently the best Cloud Storage Utilising Tool on the market in China.
To use it as a private cloud platform, privacy, security and confidentiality is a basic.
So this project should have below extra features:

- [x] Vault: a transparent encryption driver, user can easily, safely store encrypted data on remote storage provider without noticing the encryption/decrytion
  - [x] Optional: it's also compatible with [Rclone Crypt](https://rclone.org/crypt/). It's easy to manipulate the encrypted data.
  - [x] directory and filename encryption
- [x] obfuscate sensitive information in config files
- [x] server-side encryption mode (server encrypt & decrypt all data, all data flows thru server)
- [ ] client side encryption mode (client browser request data from remote storage cloud then decrypt data in client)
  - [ ] IOS/Mac: Quantumult X/Surge/ShadowRocket Box.js script. Hijack access to all Cloud Storage, also alist server will always return 302 redirect, so when app accessing alist, redirecting to storage Provider, and the header is changed by the script. any app will work with the webdav.
  - [ ] Android/Mac/Windows/Linux/Docker : Mini local proxy for Alist server, so any app/client with in the lan/device will directly access 
- [ ] auto retry in background Task: Move, Copy operations.
  - Task resume from last failed point
  - Task auto retry on failure
  - Task size validates on finish
- [ ] rapid-upload support move between local/smb and net drive with SHA1 hash return

Other features please refer to [Alist](https://github.com/alist-org/alist)

## Client Workflow

|                                       |                         |                                            |                        |                                   |                    |
| ------------------------------------- | ----------------------- | ------------------------------------------ | ---------------------- | --------------------------------- | ------------------ |
| Platform                              | Speed                   | Dependency                                 | User Step              | Compatibility                     | Available?         |
| IOS/Mac                               | Max                     | Quantumult X/Surge/ShadowRocket [Paid App] | buy App, install Rules | Browser, Any app work with webdav | work in progress   |
| Android                               | Max                     | Client App running in the background       | install apk file       | Browser, Any app work with webdav |                    |
| Mac/Windows/Linux/Docker              | Max                     | Application running in the background      | run/deploy a program   | Browser, Any app work with webdav |                    |
| Web access                            | Limited by server       |                                            | \-                     | Browser                           | Yes                |
| Web access[server set to Hybrid mode] | Max                     |                                            | \-                     | Browser                           |                    |
| Web access with alist-proxy           | Limited by proxy server | server with high bandwith, domain          | deploy alist-proxy...  | Browser, Any app work with webdav | will not implement |

## License

The `AList-Private-Cloud` is open-source software licensed under the AGPL-3.0 license.

## Last but not least
- This fork is a free and open source project. 
- 

---
