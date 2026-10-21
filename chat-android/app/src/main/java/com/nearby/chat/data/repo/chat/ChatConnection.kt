package com.nearby.chat.data.repo.chat

import com.nearby.chat.data.remote.ws.WsClient
import com.nearby.chat.data.remote.ws.WsClientMessage
import com.nearby.chat.data.remote.ws.WsServerMessage
import com.nearby.chat.data.remote.ws.WsTypes
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.SharedFlow
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ChatConnection @Inject constructor(
    private val wsClient: WsClient,
) {
    val messages: SharedFlow<WsServerMessage> = wsClient.messages

    @Volatile
    private var lastLat: Double? = null

    @Volatile
    private var lastLng: Double? = null

    fun setOnConnected(handler: () -> Unit) = wsClient.setOnConnected(handler)

    fun setTokenRefreshHandler(handler: suspend () -> Boolean) = wsClient.setTokenExpiredHandler(handler)

    fun connect() = wsClient.connect()

    fun disconnect() = wsClient.disconnect()

    fun isConnected(): Boolean = wsClient.isConnected()

    fun send(msg: WsClientMessage): Boolean = wsClient.send(msg)

    fun ensureConnected(): Boolean {
        if (wsClient.isConnected()) return true
        wsClient.retryNow()
        return false
    }

    suspend fun awaitConnected(timeoutMs: Long = 8_000): Boolean {
        if (wsClient.isConnected()) return true
        wsClient.retryNow()
        val deadline = System.currentTimeMillis() + timeoutMs
        while (System.currentTimeMillis() < deadline) {
            delay(150)
            if (wsClient.isConnected()) return true
        }
        return wsClient.isConnected()
    }

    fun onAppForeground() {
        if (!isConnected()) {
            wsClient.retryNow()
            return
        }
        val pingOk = wsClient.send(WsClientMessage(type = WsTypes.PING))
        if (!pingOk) {
            wsClient.forceReconnect()
            return
        }
        val lat = lastLat
        val lng = lastLng
        if (lat != null && lng != null) {
            sendLocation(lat, lng)
        }
    }

    fun lastLocationOrNull(): Pair<Double, Double>? {
        val lat = lastLat
        val lng = lastLng
        return if (lat != null && lng != null) lat to lng else null
    }

    fun rememberLocation(lat: Double, lng: Double) {
        lastLat = lat
        lastLng = lng
    }

    fun clearLocation() {
        lastLat = null
        lastLng = null
    }

    fun resolveCoords(lat: Double?, lng: Double?): Pair<Double?, Double?> {
        if (lat != null && lng != null) return lat to lng
        return lastLat to lastLng
    }

    fun sendLocation(lat: Double, lng: Double): Boolean = wsClient.send(
        WsClientMessage(
            type = WsTypes.LOCATION,
            lat = lat,
            lng = lng,
        ),
    )
}
