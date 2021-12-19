const HOST = "YOUR_HOST";
const TOKEN = "YOUR_TOKEN";

addEventListener("fetch", (event) => {
    const request = event.request;
    const url = new URL(request.url);
    const sign = url.searchParams.get("sign");
    if (request.method === "OPTIONS") {
        // Handle CORS preflight requests
        event.respondWith(handleOptions(request));
    } else if (sign && sign.length === 16) {
        // Handle requests to the Down server
        event.respondWith(handleDownload(request));
    } else {
        // Handle requests to the API server
        event.respondWith(handleRequest(event));
    }
});

const corsHeaders = {
    "Access-Control-Allow-Origin": "*",
    "Access-Control-Allow-Methods": "GET,HEAD,POST,OPTIONS",
    "Access-Control-Max-Age": "86400",
};

async function handleDownload(request) {
    const origin = request.headers.get("origin");
    const url = new URL(request.url);
    const path = decodeURI(url.pathname);
    const sign = url.searchParams.get("sign");
    const name = path.split("/").pop();
    const right = md5(`alist-${TOKEN}-${name}`).slice(8, 24);
    if (sign !== right) {
        const resp = new Response(
            JSON.stringify({
                code: 401,
                message: `sign mismatch`,
            }),
            {
                headers: {
                    "content-type": "application/json;charset=UTF-8",
                },
            }
        );
        resp.headers.set("Access-Control-Allow-Origin", origin);
        return resp;
    }

    let resp = await fetch(`${HOST}/api/admin/link`, {
        method: "POST",
        headers: {
            "content-type": "application/json;charset=UTF-8",
            Authorization: TOKEN,
        },
        body: JSON.stringify({
            path: path,
        }),
    });
    let res = await resp.json();
    if (res.code !== 200) {
        return new Response(JSON.stringify(res));
    }
    request = new Request(res.data.url, request);
    if (res.data.headers) {
        for (const header of res.data.headers) {
            request.headers.set(header.name, header.value);
        }
    }
    let response = await fetch(request);

    // Recreate the response so we can modify the headers
    response = new Response(response.body, response);

    // Set CORS headers
    response.headers.set("Access-Control-Allow-Origin", origin);

    // Append to/Add Vary header so browser will cache response correctly
    response.headers.append("Vary", "Origin");

    return response;
}

/**
 * Respond to the request
 * @param {Request} request
 */
async function handleRequest(event) {
    const { request } = event;

    //请求头部、返回对象
    let reqHeaders = new Headers(request.headers),
        outBody,
        outStatus = 200,
        outStatusText = "OK",
        outCt = null,
        outHeaders = new Headers({
            "Access-Control-Allow-Origin": reqHeaders.get("Origin"),
            "Access-Control-Allow-Methods": "GET, POST, PUT, PATCH, DELETE, OPTIONS",
            "Access-Control-Allow-Headers":
                reqHeaders.get("Access-Control-Allow-Headers") ||
                "Accept, Authorization, Cache-Control, Content-Type, DNT, If-Modified-Since, Keep-Alive, Origin, User-Agent, X-Requested-With, Token, x-access-token, Notion-Version",
        });

    try {
        //取域名第一个斜杠后的所有信息为代理链接
        let url = request.url.substr(8);
        url = decodeURIComponent(url.substr(url.indexOf("/") + 1));

        //需要忽略的代理
        if (
            request.method == "OPTIONS" &&
            reqHeaders.has("access-control-request-headers")
        ) {
            //输出提示
            return new Response(null, PREFLIGHT_INIT);
        } else if (
            url.length < 3 ||
            url.indexOf(".") == -1 ||
            url == "favicon.ico" ||
            url == "robots.txt"
        ) {
            return Response.redirect("https://baidu.com", 301);
        }
        //阻断
        else if (blocker.check(url)) {
            return Response.redirect("https://baidu.com", 301);
        } else {
            //补上前缀 http://
            url = url
                .replace(/https:(\/)*/, "https://")
                .replace(/http:(\/)*/, "http://");
            if (url.indexOf("://") == -1) {
                url = "http://" + url;
            }
            //构建 fetch 参数
            let fp = {
                method: request.method,
                headers: {},
            };

            //保留头部其它信息
            let he = reqHeaders.entries();
            for (let h of he) {
                if (!["content-length"].includes(h[0])) {
                    fp.headers[h[0]] = h[1];
                }
            }
            // 是否带 body
            if (["POST", "PUT", "PATCH", "DELETE"].indexOf(request.method) >= 0) {
                const ct = (reqHeaders.get("content-type") || "").toLowerCase();
                if (ct.includes("application/json")) {
                    let requestJSON = await request.json();
                    console.log(typeof requestJSON);
                    fp.body = JSON.stringify(requestJSON);
                } else if (
                    ct.includes("application/text") ||
                    ct.includes("text/html")
                ) {
                    fp.body = await request.text();
                } else if (ct.includes("form")) {
                    fp.body = await request.formData();
                } else {
                    fp.body = await request.blob();
                }
            }
            // 发起 fetch
            let fr = await fetch(url, fp);
            outCt = fr.headers.get("content-type");
            if (outCt.includes("application/text") || outCt.includes("text/html")) {
                try {
                    // 添加base
                    let newFr = new HTMLRewriter()
                        .on("head", {
                            element(element) {
                                element.prepend(`<base href="${url}" />`, {
                                    html: true,
                                });
                            },
                        })
                        .transform(fr);
                    fr = newFr;
                } catch (e) {}
            }
            outStatus = fr.status;
            outStatusText = fr.statusText;
            outBody = fr.body;
        }
    } catch (err) {
        outCt = "application/json";
        outBody = JSON.stringify({
            code: -1,
            msg: JSON.stringify(err.stack) || err,
        });
    }

    //设置类型
    if (outCt && outCt != "") {
        outHeaders.set("content-type", outCt);
    }

    let response = new Response(outBody, {
        status: outStatus,
        statusText: outStatusText,
        headers: outHeaders,
    });

    return response;
}

const blocker = {
    keys: [],
    check: function (url) {
        url = url.toLowerCase();
        let len = blocker.keys.filter((x) => url.includes(x)).length;
        return len != 0;
    },
};

function handleOptions(request) {
    // Make sure the necessary headers are present
    // for this to be a valid pre-flight request
    let headers = request.headers;
    if (
        headers.get("Origin") !== null &&
        headers.get("Access-Control-Request-Method") !== null
        // && headers.get("Access-Control-Request-Headers") !== null
    ) {
        // Handle CORS pre-flight request.
        // If you want to check or reject the requested method + headers
        // you can do that here.
        let respHeaders = {
            ...corsHeaders,
            // Allow all future content Request headers to go back to browser
            // such as Authorization (Bearer) or X-Client-Name-Version
            "Access-Control-Allow-Headers": request.headers.get(
                "Access-Control-Request-Headers"
            ),
        };

        return new Response(null, {
            headers: respHeaders,
        });
    } else {
        // Handle standard OPTIONS request.
        // If you want to allow other HTTP Methods, you can do that here.
        return new Response(null, {
            headers: {
                Allow: "GET, HEAD, POST, OPTIONS",
            },
        });
    }
}

!(function (a) {
    "use strict";
    function b(a, b) {
        var c = (65535 & a) + (65535 & b),
            d = (a >> 16) + (b >> 16) + (c >> 16);
        return (d << 16) | (65535 & c);
    }
    function c(a, b) {
        return (a << b) | (a >>> (32 - b));
    }
    function d(a, d, e, f, g, h) {
        return b(c(b(b(d, a), b(f, h)), g), e);
    }
    function e(a, b, c, e, f, g, h) {
        return d((b & c) | (~b & e), a, b, f, g, h);
    }
    function f(a, b, c, e, f, g, h) {
        return d((b & e) | (c & ~e), a, b, f, g, h);
    }
    function g(a, b, c, e, f, g, h) {
        return d(b ^ c ^ e, a, b, f, g, h);
    }
    function h(a, b, c, e, f, g, h) {
        return d(c ^ (b | ~e), a, b, f, g, h);
    }
    function i(a, c) {
        (a[c >> 5] |= 128 << c % 32), (a[(((c + 64) >>> 9) << 4) + 14] = c);
        var d,
            i,
            j,
            k,
            l,
            m = 1732584193,
            n = -271733879,
            o = -1732584194,
            p = 271733878;
        for (d = 0; d < a.length; d += 16)
            (i = m),
                (j = n),
                (k = o),
                (l = p),
                (m = e(m, n, o, p, a[d], 7, -680876936)),
                (p = e(p, m, n, o, a[d + 1], 12, -389564586)),
                (o = e(o, p, m, n, a[d + 2], 17, 606105819)),
                (n = e(n, o, p, m, a[d + 3], 22, -1044525330)),
                (m = e(m, n, o, p, a[d + 4], 7, -176418897)),
                (p = e(p, m, n, o, a[d + 5], 12, 1200080426)),
                (o = e(o, p, m, n, a[d + 6], 17, -1473231341)),
                (n = e(n, o, p, m, a[d + 7], 22, -45705983)),
                (m = e(m, n, o, p, a[d + 8], 7, 1770035416)),
                (p = e(p, m, n, o, a[d + 9], 12, -1958414417)),
                (o = e(o, p, m, n, a[d + 10], 17, -42063)),
                (n = e(n, o, p, m, a[d + 11], 22, -1990404162)),
                (m = e(m, n, o, p, a[d + 12], 7, 1804603682)),
                (p = e(p, m, n, o, a[d + 13], 12, -40341101)),
                (o = e(o, p, m, n, a[d + 14], 17, -1502002290)),
                (n = e(n, o, p, m, a[d + 15], 22, 1236535329)),
                (m = f(m, n, o, p, a[d + 1], 5, -165796510)),
                (p = f(p, m, n, o, a[d + 6], 9, -1069501632)),
                (o = f(o, p, m, n, a[d + 11], 14, 643717713)),
                (n = f(n, o, p, m, a[d], 20, -373897302)),
                (m = f(m, n, o, p, a[d + 5], 5, -701558691)),
                (p = f(p, m, n, o, a[d + 10], 9, 38016083)),
                (o = f(o, p, m, n, a[d + 15], 14, -660478335)),
                (n = f(n, o, p, m, a[d + 4], 20, -405537848)),
                (m = f(m, n, o, p, a[d + 9], 5, 568446438)),
                (p = f(p, m, n, o, a[d + 14], 9, -1019803690)),
                (o = f(o, p, m, n, a[d + 3], 14, -187363961)),
                (n = f(n, o, p, m, a[d + 8], 20, 1163531501)),
                (m = f(m, n, o, p, a[d + 13], 5, -1444681467)),
                (p = f(p, m, n, o, a[d + 2], 9, -51403784)),
                (o = f(o, p, m, n, a[d + 7], 14, 1735328473)),
                (n = f(n, o, p, m, a[d + 12], 20, -1926607734)),
                (m = g(m, n, o, p, a[d + 5], 4, -378558)),
                (p = g(p, m, n, o, a[d + 8], 11, -2022574463)),
                (o = g(o, p, m, n, a[d + 11], 16, 1839030562)),
                (n = g(n, o, p, m, a[d + 14], 23, -35309556)),
                (m = g(m, n, o, p, a[d + 1], 4, -1530992060)),
                (p = g(p, m, n, o, a[d + 4], 11, 1272893353)),
                (o = g(o, p, m, n, a[d + 7], 16, -155497632)),
                (n = g(n, o, p, m, a[d + 10], 23, -1094730640)),
                (m = g(m, n, o, p, a[d + 13], 4, 681279174)),
                (p = g(p, m, n, o, a[d], 11, -358537222)),
                (o = g(o, p, m, n, a[d + 3], 16, -722521979)),
                (n = g(n, o, p, m, a[d + 6], 23, 76029189)),
                (m = g(m, n, o, p, a[d + 9], 4, -640364487)),
                (p = g(p, m, n, o, a[d + 12], 11, -421815835)),
                (o = g(o, p, m, n, a[d + 15], 16, 530742520)),
                (n = g(n, o, p, m, a[d + 2], 23, -995338651)),
                (m = h(m, n, o, p, a[d], 6, -198630844)),
                (p = h(p, m, n, o, a[d + 7], 10, 1126891415)),
                (o = h(o, p, m, n, a[d + 14], 15, -1416354905)),
                (n = h(n, o, p, m, a[d + 5], 21, -57434055)),
                (m = h(m, n, o, p, a[d + 12], 6, 1700485571)),
                (p = h(p, m, n, o, a[d + 3], 10, -1894986606)),
                (o = h(o, p, m, n, a[d + 10], 15, -1051523)),
                (n = h(n, o, p, m, a[d + 1], 21, -2054922799)),
                (m = h(m, n, o, p, a[d + 8], 6, 1873313359)),
                (p = h(p, m, n, o, a[d + 15], 10, -30611744)),
                (o = h(o, p, m, n, a[d + 6], 15, -1560198380)),
                (n = h(n, o, p, m, a[d + 13], 21, 1309151649)),
                (m = h(m, n, o, p, a[d + 4], 6, -145523070)),
                (p = h(p, m, n, o, a[d + 11], 10, -1120210379)),
                (o = h(o, p, m, n, a[d + 2], 15, 718787259)),
                (n = h(n, o, p, m, a[d + 9], 21, -343485551)),
                (m = b(m, i)),
                (n = b(n, j)),
                (o = b(o, k)),
                (p = b(p, l));
        return [m, n, o, p];
    }
    function j(a) {
        var b,
            c = "";
        for (b = 0; b < 32 * a.length; b += 8)
            c += String.fromCharCode((a[b >> 5] >>> b % 32) & 255);
        return c;
    }
    function k(a) {
        var b,
            c = [];
        for (c[(a.length >> 2) - 1] = void 0, b = 0; b < c.length; b += 1) c[b] = 0;
        for (b = 0; b < 8 * a.length; b += 8)
            c[b >> 5] |= (255 & a.charCodeAt(b / 8)) << b % 32;
        return c;
    }
    function l(a) {
        return j(i(k(a), 8 * a.length));
    }
    function m(a, b) {
        var c,
            d,
            e = k(a),
            f = [],
            g = [];
        for (
            f[15] = g[15] = void 0, e.length > 16 && (e = i(e, 8 * a.length)), c = 0;
            16 > c;
            c += 1
        )
            (f[c] = 909522486 ^ e[c]), (g[c] = 1549556828 ^ e[c]);
        return (d = i(f.concat(k(b)), 512 + 8 * b.length)), j(i(g.concat(d), 640));
    }
    function n(a) {
        var b,
            c,
            d = "0123456789abcdef",
            e = "";
        for (c = 0; c < a.length; c += 1)
            (b = a.charCodeAt(c)), (e += d.charAt((b >>> 4) & 15) + d.charAt(15 & b));
        return e;
    }
    function o(a) {
        return unescape(encodeURIComponent(a));
    }
    function p(a) {
        return l(o(a));
    }
    function q(a) {
        return n(p(a));
    }
    function r(a, b) {
        return m(o(a), o(b));
    }
    function s(a, b) {
        return n(r(a, b));
    }
    function t(a, b, c) {
        return b ? (c ? r(b, a) : s(b, a)) : c ? p(a) : q(a);
    }
    "function" == typeof define && define.amd
        ? define(function () {
            return t;
        })
        : (a.md5 = t);
})(this);
