package com.nearby.chat.data.remote.ws

import com.nearby.chat.BuildConfig
import com.nearby.chat.core.network.ClientHeaders
import com.nearby.chat.data.local.DeviceId
import com.nearby.chat.data.local.TokenStore
import com.nearby.chat.di.ApplicationScope
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import okhttp3.HttpUrl.Companion.toHttpUrl
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import java.util.UUID
import java.util.concurrent.atomic.AtomicBoolean
import java.util.concurrent.atomic.AtomicInteger
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class WsClient @Inject constructor(
    private val okHttpClient: OkHttpClient,
    private val tokenStore: TokenStore,
    private val deviceId: DeviceId,
    private val json: Json,
    @ApplicationScope private val scope: CoroutineScope,
) {
    private var webSocket: WebSocket? = null
    private var pingJob: Job? = null
    private var reconnectJob: Job? = null
    private val manualClose = AtomicBoolean(false)
    private val reconnectAttempt = AtomicInteger(0)
    private val openMutex = Mutex()

    private val connectionState = MutableStateFlow(WsConnectionState.DISCONNECTED)

    private val _messages = MutableSharedFlow<WsServerMessage>(extraBufferCapacity = 64)
    val messages: SharedFlow<WsServerMessage> = _messages.asSharedFlow()

    private val inbound = Channel<WsServerMessage>(capacity = Channel.BUFFERED)

    private var onTokenExpired: (suspend () -> Boolean)? = null

    private var onConnected: (() -> Unit)? = null

    init {
        scope.launch {
            for (msg in inbound) {
                _messages.emit(msg)
            }
        }
    }

    fun setTokenExpiredHandler(handler: suspend () -> Boolean) {
        onTokenExpired = handler
    }

    fun setOnConnected(handler: () -> Unit) {
        onConnected = handler
    }

    fun isConnected(): Boolean = connectionState.value == WsConnectionState.CONNECTED

    fun connect() {
        manualClose.set(false)
        scope.launch { open() }
    }

    fun disconnect() {
        manualClose.set(true)
        reconnectJob?.cancel()
        stopPing()
        val old = webSocket
        webSocket = null
        connectionState.value = WsConnectionState.DISCONNECTED
        old?.close(1000, "bye")
    }

    fun retryNow() {
        if (manualClose.get()) return
        reconnectJob?.cancel()
        reconnectAttempt.set(0)
        if (connectionState.value == WsConnectionState.CONNECTED) {
            return
        }

        if (connectionState.value == WsConnectionState.CONNECTING) {
            forceReconnect()
            return
        }
        scope.launch { open() }
    }

    fun forceReconnect() {
        if (manualClose.get()) return
        reconnectJob?.cancel()
        reconnectAttempt.set(0)
        stopPing()
        val old = webSocket
        webSocket = null
        connectionState.value = WsConnectionState.DISCONNECTED
        old?.cancel()
        scope.launch { open() }
    }

    fun send(msg: WsClientMessage): Boolean {
        val ws = webSocket ?: return false
        if (connectionState.value != WsConnectionState.CONNECTED) return false
        return ws.send(json.encodeToString(msg))
    }

    private suspend fun open() {
        openMutex.withLock {
            if (connectionState.value == WsConnectionState.CONNECTING || connectionState.value == WsConnectionState.CONNECTED) {
                return
            }
            val token = tokenStore.accessToken()
            if (token.isNullOrBlank()) {
                scheduleReconnect(1000)
                return
            }
            connectionState.value = WsConnectionState.CONNECTING
            try {
                val request = Request.Builder()
                    .url(buildWsUrl())
                    .header("Authorization", "Bearer $token")
                    .build()
                val ws = okHttpClient.newWebSocket(request, listener)
                webSocket = ws
            } catch (t: Throwable) {
                webSocket = null
                connectionState.value = WsConnectionState.DISCONNECTED
                if (!manualClose.get()) scheduleReconnect()
            }
        }
    }

    private suspend fun buildWsUrl(): String {
        val base = BuildConfig.API_BASE_URL.trimEnd('/')
        val httpBase = when {
            base.startsWith("ws://") -> "http://" + base.removePrefix("ws://")
            base.startsWith("wss://") -> "https://" + base.removePrefix("wss://")
            else -> base
        }

        val httpUrl = httpBase.toHttpUrl().newBuilder()
            .addPathSegments("api/ws")
            .addQueryParameter("requestId", UUID.randomUUID().toString())
            .addQueryParameter("clientPlatform", ClientHeaders.PLATFORM_ANDROID)
            .addQueryParameter("clientVersion", BuildConfig.VERSION_NAME)
            .addQueryParameter("deviceId", deviceId.get())
            .build()
            .toString()
        return when {
            httpUrl.startsWith("https://") -> "wss://" + httpUrl.removePrefix("https://")
            httpUrl.startsWith("http://") -> "ws://" + httpUrl.removePrefix("http://")
            else -> httpUrl
        }
    }

    private val listener = object : WebSocketListener() {
        override fun onOpen(webSocket: WebSocket, response: Response) {
            if (this@WsClient.webSocket !== webSocket) return
            reconnectAttempt.set(0)
            connectionState.value = WsConnectionState.CONNECTED
            startPing()
            runCatching { onConnected?.invoke() }
        }

        override fun onMessage(webSocket: WebSocket, text: String) {
            if (this@WsClient.webSocket !== webSocket) return
            runCatching {
                json.decodeFromString(WsServerMessage.serializer(), text)
            }.onSuccess { msg ->
                inbound.trySend(msg)
            }
        }

        override fun onClosing(webSocket: WebSocket, code: Int, reason: String) {
            if (this@WsClient.webSocket !== webSocket) return
            webSocket.close(code, reason)
        }

        override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
            if (this@WsClient.webSocket !== webSocket) return
            stopPing()
            this@WsClient.webSocket = null
            connectionState.value = WsConnectionState.DISCONNECTED
            if (manualClose.get()) return
            if (code == 4401) {
                scope.launch {
                    val ok = onTokenExpired?.invoke() == true
                    if (ok) forceReconnect() else disconnect()
                }
                return
            }
            scheduleReconnect()
        }

        override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
            if (this@WsClient.webSocket !== webSocket) return
            stopPing()
            this@WsClient.webSocket = null
            connectionState.value = WsConnectionState.DISCONNECTED
            if (!manualClose.get()) scheduleReconnect()
        }
    }

    private fun startPing() {
        stopPing()
        pingJob = scope.launch {
            while (isActive) {
                delay(25_000)
                send(WsClientMessage(type = WsTypes.PING))
            }
        }
    }

    private fun stopPing() {
        pingJob?.cancel()
        pingJob = null
    }

    private fun scheduleReconnect(delayOverride: Long? = null) {
        if (manualClose.get()) return
        reconnectJob?.cancel()
        val attempt = reconnectAttempt.getAndIncrement()
        val delayMs = delayOverride ?: minOf(1000L * (1L shl attempt.coerceAtMost(4)), 15_000L)
        reconnectJob = scope.launch {
            delay(delayMs)
            if (!manualClose.get()) open()
        }
    }
}
