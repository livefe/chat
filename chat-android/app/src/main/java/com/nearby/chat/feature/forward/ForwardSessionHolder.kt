package com.nearby.chat.feature.forward

import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ForwardSessionHolder @Inject constructor() {
    private var session: ForwardSession? = null

    @Synchronized
    fun set(value: ForwardSession) {
        session = value
    }

    @Synchronized
    fun peek(): ForwardSession? = session

    @Synchronized
    fun clear(value: ForwardSession) {
        if (session == value) session = null
    }
}
