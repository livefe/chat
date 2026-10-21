package com.nearby.chat.data.repo

import android.content.Context
import com.nearby.chat.core.chat.MessageScopes
import com.nearby.chat.core.chat.messagePreview
import com.nearby.chat.core.model.ChatEvent
import com.nearby.chat.core.model.MessageItem
import com.nearby.chat.data.local.ChatLocationStore
import com.nearby.chat.data.local.TokenStore
import com.nearby.chat.data.mapper.toChatEvent
import com.nearby.chat.data.mapper.toDomain
import com.nearby.chat.data.remote.ws.WsScopes
import com.nearby.chat.data.remote.ws.WsTypes
import com.nearby.chat.data.remote.ws.messageItem
import com.nearby.chat.data.repo.chat.ChatConnection
import com.nearby.chat.data.repo.chat.ChatSendPipeline
import com.nearby.chat.data.repo.chat.ChatSubscriptions
import com.nearby.chat.data.repo.chat.ChatUnreadTracker
import com.nearby.chat.di.ApplicationScope
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.mapNotNull
import kotlinx.coroutines.launch
import java.util.concurrent.ConcurrentHashMap
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ChatSession @Inject constructor(
    private val connection: ChatConnection,
    private val subscriptions: ChatSubscriptions,
    private val sendPipeline: ChatSendPipeline,
    private val unread: ChatUnreadTracker,
    private val chatLocationStore: ChatLocationStore,
    private val tokenStore: TokenStore,
    @ApplicationScope private val scope: CoroutineScope,
    @ApplicationContext private val context: Context,
) {
    val events: Flow<ChatEvent> = connection.messages.mapNotNull { it.toChatEvent() }

    private val _reconnectTick = MutableSharedFlow<Unit>(extraBufferCapacity = 1)
    val reconnectTick: SharedFlow<Unit> = _reconnectTick.asSharedFlow()

    private val privateClearedBefore = ConcurrentHashMap<Long, Long>()

    init {
        connection.setOnConnected { onWsConnected() }
        scope.launch {
            connection.messages.collect { msg ->
                subscriptions.resolvePending(msg)
                sendPipeline.resolvePending(msg)
                if (msg.type != WsTypes.MESSAGE) return@collect
                val item = msg.messageItem()?.toDomain() ?: return@collect
                val isPrivate =
                    msg.scope == WsScopes.PRIVATE || item.scope == MessageScopes.PRIVATE
                if (!isPrivate) return@collect
                val selfId = tokenStore.user()?.id ?: 0L
                if (selfId <= 0L || item.fromUserId == selfId) return@collect
                if (!messageVisible(item.fromUserId, item)) return@collect
                unread.onIncomingPrivate(
                    fromUserId = item.fromUserId,
                    title = item.fromName,
                    body = messagePreview(item.contentType, item.content).resolve(context),
                )
            }
        }
    }

    fun setPrivateClearedBefore(peerId: Long, clearedBeforeId: Long) {
        if (peerId <= 0L || clearedBeforeId <= 0L) return
        privateClearedBefore.merge(peerId, clearedBeforeId) { a, b -> maxOf(a, b) }
    }

    fun messageVisible(peerId: Long, message: MessageItem): Boolean {
        if (message.id == 0L) return true
        return message.id > (privateClearedBefore[peerId] ?: 0L)
    }

    fun setTokenRefreshHandler(handler: suspend () -> Boolean) = connection.setTokenRefreshHandler(handler)

    fun connect() = connection.connect()

    fun disconnect() {
        subscriptions.clear()
        privateClearedBefore.clear()
        sendPipeline.clear()
        unread.clear()
        connection.clearLocation()
        connection.disconnect()
    }

    private fun onWsConnected() {
        scope.launch {
            var loc = connection.lastLocationOrNull()
            if (loc == null) {
                val saved = runCatching { chatLocationStore.load() }.getOrNull()
                if (saved != null) {
                    connection.rememberLocation(saved.lat, saved.lng)
                    loc = saved.lat to saved.lng
                }
            }
            if (loc != null) {
                connection.sendLocation(loc.first, loc.second)
            }
            subscriptions.resubscribeAll(loc?.first, loc?.second)
            _reconnectTick.tryEmit(Unit)
        }
    }

    fun reportLocation(lat: Double, lng: Double): Boolean {
        connection.rememberLocation(lat, lng)
        if (!connection.ensureConnected()) return false
        return connection.sendLocation(lat, lng)
    }
}
