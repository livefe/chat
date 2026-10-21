package com.nearby.chat.core.format
private val maskedMobile = Regex("""^\d{3}\*{4}\d{4}$""")

fun sanitizeNickname(raw: String?): String {
    val name = raw?.trim().orEmpty()
    if (name.isEmpty()) return ""
    if (name.length == 11 && name.all { it.isDigit() }) return ""
    if (maskedMobile.matches(name)) return ""
    return name
}
