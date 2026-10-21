package com.nearby.chat.data.push

interface PushChannel {
    val vendor: PushVendor
    fun getToken(onToken: TokenListener)
}
