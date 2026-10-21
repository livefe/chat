package com.nearby.chat.feature.main

import androidx.lifecycle.ViewModel
import com.nearby.chat.core.share.ShareLinks
import com.nearby.chat.data.repo.SessionRepository
import com.nearby.chat.feature.forward.ForwardSession
import com.nearby.chat.feature.forward.ForwardSessionHolder
import com.nearby.chat.feature.navigation.resolvePendingShareTarget
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import javax.inject.Inject

@HiltViewModel
class MainViewModel @Inject constructor(
    sessionRepository: SessionRepository,
    private val forwardSessionHolder: ForwardSessionHolder,
) : ViewModel() {
    val hasToken: Flow<Boolean> = sessionRepository.hasToken

    private val _deepLink = MutableStateFlow<ShareLinks.Target?>(null)
    val deepLink: StateFlow<ShareLinks.Target?> = _deepLink.asStateFlow()

    private var handledDeepLinkKey: String? = null

    fun onDeepLink(target: ShareLinks.Target?) {
        _deepLink.value = target
    }

    fun claimDeepLink(target: ShareLinks.Target): ShareLinks.Target? {
        val resolved = resolvePendingShareTarget(target) ?: return null
        val key = ShareLinks.targetKey(resolved)
        if (key == handledDeepLinkKey) return null
        handledDeepLinkKey = key
        return resolved
    }

    fun startForward(session: ForwardSession): Boolean {
        if (session.content.isBlank()) return false
        forwardSessionHolder.set(session)
        return true
    }
}
