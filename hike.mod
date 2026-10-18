module github.com/kanryu/slint

hike 1.0

package ./internal/cflat {
    target windows {
        hike-object: "none"
        link: "deps/slint-cflat-windows/libslint_cflat.a"
        link: "-Wl,-u,slint_c_create"
        link: "-lgdi32"
        link: "-lopengl32"
        link: "-limm32"
        link: "-lntdll"
        link: "-luser32"
        link: "-lkernel32"
        link: "-lole32"
        link: "-loleaut32"
        link: "-lcomctl32"
        link: "-ldwmapi"
        link: "-ldwrite"
        link: "-lshlwapi"
        link: "-lshell32"
        link: "-luxtheme"
        link: "-lpropsys"
        link: "-lmsimg32"
        link: "-lusp10"
        link: "-lfontsub"
        link: "-lurlmon"
        link: "-loleacc"
        link: "-luiautomationcore"
        link: "-ladvapi32"
        link: "-lwinhttp"
        link: "-lws2_32"
        link: "-lucrt"
        link: "-lucrtbase"
        link: "-lm"
        link: "-lgcc"
        link: "-lgcc_eh"
        link: "-lmingwex"
        link: "-lmingw32"
    }
}
