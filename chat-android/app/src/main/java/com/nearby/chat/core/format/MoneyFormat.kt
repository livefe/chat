package com.nearby.chat.core.format
object MoneyFormat {
    fun formatAmount(v: Double): String = if (v == v.toLong().toDouble()) v.toLong().toString() else v.toString()
}
