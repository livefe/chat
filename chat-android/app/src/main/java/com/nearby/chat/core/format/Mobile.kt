package com.nearby.chat.core.format

private val cnMobileRegex = Regex("^1[3-9]\\d{9}$")

private const val CN_MOBILE_LENGTH = 11

fun isValidCnMobile(mobile: String): Boolean = cnMobileRegex.matches(mobile)

fun maskMobile(mobile: String): String {
    if (mobile.length != CN_MOBILE_LENGTH) return mobile
    return "${mobile.take(3)}****${mobile.takeLast(4)}"
}
