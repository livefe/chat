package com.nearby.chat.data.repo.chat

import android.content.Context
import com.nearby.chat.R
import com.nearby.chat.data.remote.ws.WsClientMessage
import com.nearby.chat.data.remote.ws.WsScopes
import com.nearby.chat.data.remote.ws.WsServerMessage
import com.nearby.chat.data.remote.ws.WsTypes
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.withTimeoutOrNull
import java.util.concurrent.ConcurrentHashMap
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ChatSubscriptions @Inject constructor(
    private val connection: ChatConnection,
    @ApplicationContext private val context: Context,
) {
    private val subscribedRooms = ConcurrentHashMap.newKeySet<Long>()
    private val subscribedPosts = ConcurrentHashMap.newKeySet<Long>()
    private val pendingRoomSubscribe = ConcurrentHashMap<Long, CompletableDeferred<Result<Unit>>>()
    private val pendingPostSubscribe = ConcurrentHashMap<Long, CompletableDeferred<Result<Unit>>>()

    fun resolvePending(msg: WsServerMessage) {
        when (msg.type) {
            WsTypes.ROOM_PRESENCE -> {
                val roomId = msg.roomId ?: return
                pendingRoomSubscribe.remove(roomId)?.complete(Result.success(Unit))
            }

            WsTypes.POST_PRESENCE -> {
                val postId = msg.postId ?: return
                pendingPostSubscribe.remove(postId)?.complete(Result.success(Unit))
            }

            WsTypes.ERROR -> {
                val roomId = msg.roomId ?: 0L
                if (roomId > 0) {
                    pendingRoomSubscribe.remove(roomId)
                        ?.complete(failure(msg.content ?: context.getString(R.string.chat_join_room_failed)))
                }
                val postId = msg.postId ?: 0L
                if (postId > 0) {
                    pendingPostSubscribe.remove(postId)
                        ?.complete(failure(msg.content ?: context.getString(R.string.chat_join_post_failed)))
                }
            }
        }
    }

    fun resubscribeAll(lat: Double?, lng: Double?) {
        for (roomId in subscribedRooms) {
            sendRoomSubscribe(roomId, lat, lng)
        }
        for (postId in subscribedPosts) {
            sendPostSubscribe(postId)
        }
    }

    fun clear() {
        val disconnected = context.getString(R.string.chat_disconnected)
        for (roomId in subscribedRooms.toList()) {
            pendingRoomSubscribe.remove(roomId)?.complete(failure(disconnected))
        }
        for (postId in subscribedPosts.toList()) {
            pendingPostSubscribe.remove(postId)?.complete(failure(disconnected))
        }
        subscribedRooms.clear()
        subscribedPosts.clear()
        pendingRoomSubscribe.clear()
        pendingPostSubscribe.clear()
    }

    fun isRoomSubscribed(roomId: Long): Boolean = roomId > 0 && subscribedRooms.contains(roomId)

    suspend fun awaitSubscribeRoom(
        roomId: Long,
        lat: Double? = null,
        lng: Double? = null,
        timeoutMs: Long = 10_000,
    ): Result<Unit> = awaitSubscribe(
        id = roomId,
        pending = pendingRoomSubscribe,
        timeoutMs = timeoutMs,
        invalidMessage = context.getString(R.string.chat_invalid_room),
        timeoutMessage = context.getString(R.string.chat_join_room_timeout),
        start = { subscribeRoom(roomId, lat, lng) },
        onFailure = { unsubscribeRoom(roomId) },
    )

    fun unsubscribeRoom(roomId: Long) {
        if (roomId <= 0) return
        pendingRoomSubscribe.remove(roomId)?.cancel()
        subscribedRooms.remove(roomId)
        if (!connection.isConnected()) return
        connection.send(
            WsClientMessage(
                type = WsTypes.UNSUBSCRIBE,
                scope = WsScopes.ROOM,
                roomId = roomId,
            ),
        )
    }

    suspend fun awaitSubscribePost(postId: Long, timeoutMs: Long = 10_000): Result<Unit> = awaitSubscribe(
        id = postId,
        pending = pendingPostSubscribe,
        timeoutMs = timeoutMs,
        invalidMessage = context.getString(R.string.chat_invalid_post),
        timeoutMessage = context.getString(R.string.chat_join_post_timeout),
        start = { subscribePost(postId) },
        onFailure = { unsubscribePost(postId) },
    )

    fun unsubscribePost(postId: Long) {
        if (postId <= 0) return
        pendingPostSubscribe.remove(postId)?.cancel()
        subscribedPosts.remove(postId)
        if (!connection.isConnected()) return
        connection.send(
            WsClientMessage(
                type = WsTypes.UNSUBSCRIBE,
                scope = WsScopes.POST,
                postId = postId,
            ),
        )
    }

    private fun subscribeRoom(roomId: Long, lat: Double?, lng: Double?) {
        if (roomId <= 0) return
        subscribedRooms.add(roomId)
        val (useLat, useLng) = connection.resolveCoords(lat, lng)
        if (!connection.ensureConnected()) return
        sendRoomSubscribe(roomId, useLat, useLng)
    }

    private fun subscribePost(postId: Long) {
        if (postId <= 0) return
        subscribedPosts.add(postId)
        if (!connection.ensureConnected()) return
        sendPostSubscribe(postId)
    }

    private suspend fun awaitSubscribe(
        id: Long,
        pending: ConcurrentHashMap<Long, CompletableDeferred<Result<Unit>>>,
        timeoutMs: Long,
        invalidMessage: String,
        timeoutMessage: String,
        start: () -> Unit,
        onFailure: () -> Unit,
    ): Result<Unit> {
        if (id <= 0) return Result.failure(IllegalArgumentException(invalidMessage))
        if (!connection.awaitConnected()) {
            return failure(context.getString(R.string.chat_connect_failed_retry))
        }
        val deferred = CompletableDeferred<Result<Unit>>()
        val previous = pending.put(id, deferred)
        if (previous != null && previous !== deferred) {
            previous.complete(failure(context.getString(R.string.chat_subscription_switched)))
        }
        start()
        val result = withTimeoutOrNull(timeoutMs) { deferred.await() }
            ?: failure(timeoutMessage)
        pending.remove(id, deferred)
        if (result.isFailure) onFailure()
        return result
    }

    private fun sendRoomSubscribe(roomId: Long, lat: Double?, lng: Double?) {
        connection.send(
            WsClientMessage(
                type = WsTypes.SUBSCRIBE,
                scope = WsScopes.ROOM,
                roomId = roomId,
                lat = lat,
                lng = lng,
            ),
        )
    }

    private fun sendPostSubscribe(postId: Long) {
        connection.send(
            WsClientMessage(
                type = WsTypes.SUBSCRIBE,
                scope = WsScopes.POST,
                postId = postId,
            ),
        )
    }

    private fun failure(message: String): Result<Unit> = Result.failure(IllegalStateException(message))
}
