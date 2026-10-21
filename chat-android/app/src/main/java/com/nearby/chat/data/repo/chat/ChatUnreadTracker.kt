package com.nearby.chat.data.repo.chat

import com.nearby.chat.core.network.ensureSuccess
import com.nearby.chat.core.network.unwrap
import com.nearby.chat.data.push.InAppMessageNotifier
import com.nearby.chat.data.remote.api.ConversationApi
import com.nearby.chat.data.remote.dto.PeerIdRequest
import com.nearby.chat.di.ApplicationScope
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ChatUnreadTracker @Inject constructor(
    private val conversationApi: ConversationApi,
    private val notifier: InAppMessageNotifier,
    @ApplicationScope private val scope: CoroutineScope,
) {
    private val _unreadByPeer = MutableStateFlow<Map<Long, Int>>(emptyMap())
    val unreadByPeer: StateFlow<Map<Long, Int>> = _unreadByPeer.asStateFlow()

    private val _unreadTotal = MutableStateFlow(0)
    val unreadTotal: StateFlow<Int> = _unreadTotal.asStateFlow()

    @Volatile
    private var activePrivatePeerId: Long = 0

    fun onIncomingPrivate(fromUserId: Long, title: String, body: String) {
        if (fromUserId == activePrivatePeerId) {
            clearPeerUnreadLocal(fromUserId)
            notifier.cancelPrivate(fromUserId)
            markReadAsync(fromUserId)
            return
        }
        incrPeerUnreadLocal(fromUserId)
        notifier.notifyPrivate(fromUserId, title, body)
    }

    fun setActivePrivatePeer(peerId: Long) {
        activePrivatePeerId = peerId
        if (peerId > 0) notifier.cancelPrivate(peerId)
    }

    fun clearActivePrivatePeer(peerId: Long) {
        if (activePrivatePeerId == peerId) {
            activePrivatePeerId = 0
        }
    }

    fun clear() {
        _unreadByPeer.value = emptyMap()
        _unreadTotal.value = 0
        activePrivatePeerId = 0
    }

    suspend fun refreshUnread() {
        runCatching { conversationApi.unread().unwrap() }
            .onSuccess { summary ->
                val map = summary.items
                    .filter { it.count > 0 && it.peerId > 0L }
                    .associate { it.peerId to it.count }
                _unreadByPeer.value = map
                _unreadTotal.value = summary.total.coerceAtLeast(0)
            }
    }

    suspend fun markRead(peerId: Long) {
        if (peerId <= 0L) return
        val hadLocal = (_unreadByPeer.value[peerId] ?: 0) > 0
        clearPeerUnreadLocal(peerId)
        runCatching { conversationApi.markRead(PeerIdRequest(peerId)).ensureSuccess() }
        if (hadLocal) refreshUnread()
    }

    fun markReadAsync(peerId: Long) {
        if (peerId <= 0L) return
        scope.launch { markRead(peerId) }
    }

    fun clearPeerUnreadLocal(peerId: Long) {
        if (peerId <= 0L) return
        val removed = _unreadByPeer.value[peerId] ?: 0
        if (removed <= 0) return
        _unreadByPeer.update { it - peerId }
        _unreadTotal.update { (it - removed).coerceAtLeast(0) }
    }

    private fun incrPeerUnreadLocal(peerId: Long) {
        if (peerId <= 0L) return
        _unreadByPeer.update { cur ->
            cur + (peerId to ((cur[peerId] ?: 0) + 1))
        }
        _unreadTotal.update { it + 1 }
    }
}
