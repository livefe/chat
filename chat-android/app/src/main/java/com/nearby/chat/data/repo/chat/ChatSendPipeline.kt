package com.nearby.chat.data.repo.chat

import android.content.Context
import com.nearby.chat.R
import com.nearby.chat.core.chat.ChatLimits
import com.nearby.chat.core.model.SendFailedEvent
import com.nearby.chat.data.remote.ws.WsClientMessage
import com.nearby.chat.data.remote.ws.WsScopes
import com.nearby.chat.data.remote.ws.WsServerMessage
import com.nearby.chat.data.remote.ws.WsTypes
import com.nearby.chat.di.ApplicationScope
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.launch
import java.util.concurrent.ConcurrentHashMap
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ChatSendPipeline @Inject constructor(
    private val connection: ChatConnection,
    @ApplicationScope private val scope: CoroutineScope,
    @ApplicationContext private val context: Context,
) {
    private val pendingSendIds = ConcurrentHashMap.newKeySet<String>()
    private val pendingSendJobs = ConcurrentHashMap<String, Job>()

    private val _sendFailed = MutableSharedFlow<SendFailedEvent>(extraBufferCapacity = 16)
    val sendFailed: SharedFlow<SendFailedEvent> = _sendFailed.asSharedFlow()

    fun resolvePending(msg: WsServerMessage) {
        val clientMsgId = msg.clientMsgId?.takeIf { it.isNotBlank() } ?: return
        when (msg.type) {
            WsTypes.ACK, WsTypes.MESSAGE -> clearPendingSend(clientMsgId)

            WsTypes.ERROR -> {
                if (clearPendingSend(clientMsgId)) {
                    _sendFailed.tryEmit(
                        SendFailedEvent(
                            clientMsgId = clientMsgId,
                            reason = msg.content.orEmpty(),
                        ),
                    )
                }
            }
        }
    }

    fun clear() {
        pendingSendIds.clear()
        pendingSendJobs.values.forEach { it.cancel() }
        pendingSendJobs.clear()
    }

    fun sendNearby(
        content: String,
        lat: Double,
        lng: Double,
        contentType: Int,
        clientMsgId: String,
        replyToId: Long,
    ): Boolean {
        connection.rememberLocation(lat, lng)
        if (!connection.ensureConnected()) return false
        connection.sendLocation(lat, lng)
        return sendTracked(
            clientMsgId,
            WsClientMessage(
                type = WsTypes.SEND,
                scope = WsScopes.NEARBY,
                contentType = contentType,
                content = content,
                lat = lat,
                lng = lng,
                replyToId = replyToId.takeIf { it > 0 },
                clientMsgId = clientMsgId,
            ),
        )
    }

    fun sendPrivate(
        toUserId: Long,
        content: String,
        contentType: Int,
        clientMsgId: String,
        replyToId: Long,
    ): Boolean {
        if (!connection.ensureConnected()) return false
        return sendTracked(
            clientMsgId,
            WsClientMessage(
                type = WsTypes.SEND,
                scope = WsScopes.PRIVATE,
                toUserId = toUserId,
                contentType = contentType,
                content = content,
                replyToId = replyToId.takeIf { it > 0 },
                clientMsgId = clientMsgId,
            ),
        )
    }

    fun sendRoom(
        roomId: Long,
        content: String,
        lat: Double?,
        lng: Double?,
        contentType: Int,
        clientMsgId: String,
        replyToId: Long,
    ): Boolean {
        if (!connection.ensureConnected()) return false
        val (useLat, useLng) = connection.resolveCoords(lat, lng)
        return sendTracked(
            clientMsgId,
            WsClientMessage(
                type = WsTypes.SEND,
                scope = WsScopes.ROOM,
                roomId = roomId,
                contentType = contentType,
                content = content,
                lat = useLat,
                lng = useLng,
                replyToId = replyToId.takeIf { it > 0 },
                clientMsgId = clientMsgId,
            ),
        )
    }

    fun sendPost(
        postId: Long,
        content: String,
        contentType: Int,
        clientMsgId: String,
        replyToId: Long,
    ): Boolean {
        if (!connection.ensureConnected()) return false
        return sendTracked(
            clientMsgId,
            WsClientMessage(
                type = WsTypes.SEND,
                scope = WsScopes.POST,
                postId = postId,
                contentType = contentType,
                content = content,
                replyToId = replyToId.takeIf { it > 0 },
                clientMsgId = clientMsgId,
            ),
        )
    }

    private fun sendTracked(clientMsgId: String, msg: WsClientMessage): Boolean {
        val ok = connection.send(msg)
        if (ok) trackPendingSend(clientMsgId)
        return ok
    }

    private fun trackPendingSend(clientMsgId: String) {
        if (clientMsgId.isBlank()) return
        clearPendingSend(clientMsgId)
        pendingSendIds.add(clientMsgId)
        pendingSendJobs[clientMsgId] = scope.launch {
            delay(ChatLimits.PENDING_SEND_TIMEOUT_MS)
            if (pendingSendIds.remove(clientMsgId)) {
                pendingSendJobs.remove(clientMsgId)
                _sendFailed.tryEmit(
                    SendFailedEvent(
                        clientMsgId = clientMsgId,
                        reason = context.getString(R.string.chat_send_timeout),
                    ),
                )
            }
        }
    }

    private fun clearPendingSend(clientMsgId: String): Boolean {
        val wasPending = pendingSendIds.remove(clientMsgId)
        pendingSendJobs.remove(clientMsgId)?.cancel()
        return wasPending
    }
}
