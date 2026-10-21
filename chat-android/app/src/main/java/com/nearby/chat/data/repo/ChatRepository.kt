package com.nearby.chat.data.repo

import com.nearby.chat.core.model.ChatEvent
import com.nearby.chat.core.model.ConversationItem
import com.nearby.chat.core.model.MessageItem
import com.nearby.chat.core.model.NearbyBootstrap
import com.nearby.chat.core.model.NearbyRoomItem
import com.nearby.chat.core.model.RoomApplicationItem
import com.nearby.chat.core.model.RoomBootstrap
import com.nearby.chat.core.model.RoomDetail
import com.nearby.chat.core.model.SendFailedEvent
import com.nearby.chat.core.network.unwrap
import com.nearby.chat.data.mapper.toDomain
import com.nearby.chat.data.remote.api.ConversationApi
import com.nearby.chat.data.remote.api.MessageApi
import com.nearby.chat.data.remote.api.RoomApi
import com.nearby.chat.data.remote.dto.CreateRoomRequest
import com.nearby.chat.data.remote.dto.IdRequest
import com.nearby.chat.data.remote.dto.PeerIdRequest
import com.nearby.chat.data.remote.dto.PinConversationRequest
import com.nearby.chat.data.repo.chat.ChatConnection
import com.nearby.chat.data.repo.chat.ChatSendPipeline
import com.nearby.chat.data.repo.chat.ChatSubscriptions
import com.nearby.chat.data.repo.chat.ChatUnreadTracker
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ChatRepository @Inject constructor(
    private val roomApi: RoomApi,
    private val messageApi: MessageApi,
    private val conversationApi: ConversationApi,
    private val session: ChatSession,
    private val connection: ChatConnection,
    private val subscriptions: ChatSubscriptions,
    private val sendPipeline: ChatSendPipeline,
    private val unread: ChatUnreadTracker,
) {
    val events: Flow<ChatEvent> = session.events
    val reconnectTick: SharedFlow<Unit> = session.reconnectTick
    val sendFailed: SharedFlow<SendFailedEvent> = sendPipeline.sendFailed
    val unreadByPeer: StateFlow<Map<Long, Int>> = unread.unreadByPeer
    val unreadTotal: StateFlow<Int> = unread.unreadTotal

    private val _nearbyRooms = MutableStateFlow<List<NearbyRoomItem>>(emptyList())
    val nearbyRooms: StateFlow<List<NearbyRoomItem>> = _nearbyRooms.asStateFlow()

    fun messageVisible(peerId: Long, message: MessageItem): Boolean = session.messageVisible(peerId, message)

    fun setActivePrivatePeer(peerId: Long) = unread.setActivePrivatePeer(peerId)

    fun clearActivePrivatePeer(peerId: Long) = unread.clearActivePrivatePeer(peerId)

    fun setTokenRefreshHandler(handler: suspend () -> Boolean) = session.setTokenRefreshHandler(handler)

    fun connect() = session.connect()

    fun disconnect() {
        _nearbyRooms.value = emptyList()
        session.disconnect()
    }

    fun isConnected(): Boolean = connection.isConnected()

    fun onAppForeground() = connection.onAppForeground()

    suspend fun awaitConnected(timeoutMs: Long = 8_000): Boolean = connection.awaitConnected(timeoutMs)

    fun lastLocationOrNull(): Pair<Double, Double>? = connection.lastLocationOrNull()

    fun rememberLocation(lat: Double, lng: Double) = connection.rememberLocation(lat, lng)

    suspend fun syncLocationForNearby(
        lat: Double,
        lng: Double,
        radius: Int,
        limit: Int,
        withMessages: Boolean,
        publishRooms: Boolean = false,
    ): NearbyBootstrap {
        session.reportLocation(lat, lng)
        val boot = roomApi.nearbyBootstrap(lat, lng, radius, limit, withMessages).unwrap().toDomain()
        if (publishRooms) _nearbyRooms.value = boot.rooms
        return boot
    }

    suspend fun nearbyMessages(
        lat: Double,
        lng: Double,
        beforeId: Long,
        radius: Int,
        limit: Int,
    ): List<MessageItem> = messageApi.nearbyMessages(lat, lng, beforeId, radius, limit).unwrap().list.map { it.toDomain() }

    suspend fun roomDetail(roomId: Long): RoomDetail = roomApi.roomDetail(roomId).unwrap().toDomain()

    suspend fun roomBootstrap(
        roomId: Long,
        limit: Int,
        withMessages: Boolean,
        lat: Double?,
        lng: Double?,
    ): RoomBootstrap = roomApi.roomBootstrap(roomId, limit, withMessages, lat, lng).unwrap().toDomain()

    suspend fun createRoom(
        name: String,
        placeType: Int,
        centerLat: Double,
        centerLng: Double,
        radius: Int,
        description: String,
    ) {
        roomApi.createRoom(
            CreateRoomRequest(
                name = name,
                placeType = placeType,
                centerLat = centerLat,
                centerLng = centerLng,
                radius = radius,
                description = description,
            ),
        ).unwrap()
    }

    suspend fun myRoomApplications(page: Int, pageSize: Int): List<RoomApplicationItem> = roomApi.myRoomApplications(page, pageSize).unwrap().list.map { it.toDomain() }

    suspend fun roomMessages(
        roomId: Long,
        beforeId: Long,
        limit: Int,
        lat: Double?,
        lng: Double?,
    ): List<MessageItem> = messageApi.roomMessages(roomId, beforeId, limit, lat, lng).unwrap().list.map { it.toDomain() }

    suspend fun privateMessages(
        peerId: Long,
        beforeId: Long,
        limit: Int,
    ): List<MessageItem> {
        val result = messageApi.privateMessages(peerId, beforeId, limit).unwrap()
        session.setPrivateClearedBefore(peerId, result.clearedBeforeId)
        return result.list.map { it.toDomain() }
    }

    suspend fun postMessages(
        postId: Long,
        beforeId: Long,
        limit: Int,
    ): List<MessageItem> = messageApi.postMessages(postId, beforeId, limit).unwrap().list.map { it.toDomain() }

    suspend fun conversationList(
        page: Int,
        pageSize: Int,
    ): Pair<Int, List<ConversationItem>> {
        val data = conversationApi.conversationList(page, pageSize).unwrap()
        return data.total to data.list.map { it.toDomain() }
    }

    suspend fun deleteConversation(peerId: Long) {
        if (peerId <= 0L) return
        val result = conversationApi.deleteConversation(PeerIdRequest(peerId)).unwrap()
        session.setPrivateClearedBefore(peerId, result.clearedBeforeId)
        unread.clearPeerUnreadLocal(peerId)
        unread.refreshUnread()
    }

    suspend fun pinConversation(peerId: Long, pinned: Boolean) {
        if (peerId <= 0L) return
        conversationApi.pinConversation(PinConversationRequest(peerId, pinned)).unwrap()
    }

    suspend fun recallMessage(id: Long) {
        if (id <= 0L) return
        messageApi.recallMessage(IdRequest(id)).unwrap()
    }

    suspend fun refreshUnread() = unread.refreshUnread()

    suspend fun markRead(peerId: Long) = unread.markRead(peerId)

    fun markReadAsync(peerId: Long) = unread.markReadAsync(peerId)

    fun isRoomSubscribed(roomId: Long): Boolean = subscriptions.isRoomSubscribed(roomId)

    suspend fun awaitSubscribeRoom(
        roomId: Long,
        lat: Double? = null,
        lng: Double? = null,
    ): Result<Unit> = subscriptions.awaitSubscribeRoom(roomId, lat, lng)

    fun unsubscribeRoom(roomId: Long) = subscriptions.unsubscribeRoom(roomId)

    suspend fun awaitSubscribePost(postId: Long): Result<Unit> = subscriptions.awaitSubscribePost(postId)

    fun unsubscribePost(postId: Long) = subscriptions.unsubscribePost(postId)

    fun sendNearby(
        content: String,
        lat: Double,
        lng: Double,
        contentType: Int,
        clientMsgId: String,
        replyToId: Long = 0,
    ): Boolean = sendPipeline.sendNearby(content, lat, lng, contentType, clientMsgId, replyToId)

    fun sendPrivate(
        toUserId: Long,
        content: String,
        contentType: Int,
        clientMsgId: String,
        replyToId: Long = 0,
    ): Boolean = sendPipeline.sendPrivate(toUserId, content, contentType, clientMsgId, replyToId)

    fun sendRoom(
        roomId: Long,
        content: String,
        lat: Double?,
        lng: Double?,
        contentType: Int,
        clientMsgId: String,
        replyToId: Long = 0,
    ): Boolean = sendPipeline.sendRoom(roomId, content, lat, lng, contentType, clientMsgId, replyToId)

    fun sendPost(
        postId: Long,
        content: String,
        contentType: Int,
        clientMsgId: String,
        replyToId: Long = 0,
    ): Boolean = sendPipeline.sendPost(postId, content, contentType, clientMsgId, replyToId)
}
