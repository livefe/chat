package com.nearby.chat.data.push

import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class PushTokenStore @Inject constructor() {
    private var pending: TokenListener? = null

    var lastToken: String = ""
        private set

    @Synchronized
    fun setPendingListener(listener: TokenListener) {
        pending = listener
        if (lastToken.isNotBlank()) {
            listener.onToken(lastToken)
        }
    }

    @Synchronized
    fun deliver(token: String) {
        if (token.isBlank()) return
        lastToken = token
        pending?.onToken(token)
    }
}
