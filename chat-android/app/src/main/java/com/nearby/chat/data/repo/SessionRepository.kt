package com.nearby.chat.data.repo

import com.nearby.chat.core.model.UserInfo
import com.nearby.chat.data.local.TokenStore
import kotlinx.coroutines.flow.Flow
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class SessionRepository @Inject constructor(
    private val tokenStore: TokenStore,
) {
    val hasToken: Flow<Boolean> = tokenStore.hasToken

    val userFlow: Flow<UserInfo?> = tokenStore.userFlow

    suspend fun currentUser(): UserInfo? = tokenStore.user()

    fun cachedUser(): UserInfo? = tokenStore.userOrNull()

    suspend fun selfId(): Long = tokenStore.user()?.id ?: 0L
}
