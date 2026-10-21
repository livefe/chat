package com.nearby.chat.data.push

import android.os.Build

object VendorDetector {
    fun detect(): PushVendor {
        val manufacturer = Build.MANUFACTURER.orEmpty().lowercase()
        val brand = Build.BRAND.orEmpty().lowercase()
        val key = "$manufacturer $brand"
        return when {
            "huawei" in key || "华为" in key -> PushVendor.HUAWEI
            "honor" in key || "荣耀" in key -> PushVendor.HONOR
            "xiaomi" in key || "redmi" in key || "小米" in key -> PushVendor.XIAOMI
            "oppo" in key || "realme" in key || "oneplus" in key -> PushVendor.OPPO
            "vivo" in key || "iqoo" in key -> PushVendor.VIVO
            else -> PushVendor.UNKNOWN
        }
    }
}
