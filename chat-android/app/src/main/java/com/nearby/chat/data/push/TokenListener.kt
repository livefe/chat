package com.nearby.chat.data.push

fun interface TokenListener {
    fun onToken(token: String)
}
