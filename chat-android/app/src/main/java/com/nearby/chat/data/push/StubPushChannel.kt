package com.nearby.chat.data.push

class StubPushChannel(
    override val vendor: PushVendor,
) : PushChannel {
    override fun getToken(onToken: TokenListener) {
        val slug = vendor.name.lowercase()
        onToken.onToken("stub-$slug-${System.currentTimeMillis()}")
    }
}
