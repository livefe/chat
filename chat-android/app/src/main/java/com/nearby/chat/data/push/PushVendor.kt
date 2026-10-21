package com.nearby.chat.data.push

enum class PushVendor {
    HUAWEI,
    HONOR,
    XIAOMI,
    OPPO,
    VIVO,
    UNKNOWN,
    ;

    fun apiValue(): String = when (this) {
        HUAWEI -> "huawei"
        HONOR -> "honor"
        XIAOMI -> "xiaomi"
        OPPO -> "oppo"
        VIVO -> "vivo"
        UNKNOWN -> "other"
    }
}
