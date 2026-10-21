package com.nearby.chat.data.repo

import com.nearby.chat.core.model.BlockListItem
import com.nearby.chat.core.model.PublicUserProfile
import com.nearby.chat.core.model.UserInfo
import com.nearby.chat.core.network.ensureSuccess
import com.nearby.chat.core.network.unwrap
import com.nearby.chat.data.local.ChatLocationStore
import com.nearby.chat.data.local.TokenStore
import com.nearby.chat.data.mapper.toDomain
import com.nearby.chat.data.network.TokenRefreshCoordinator
import com.nearby.chat.data.remote.api.UserApi
import com.nearby.chat.data.remote.dto.BlockUserRequest
import com.nearby.chat.data.remote.dto.ChangeMobileRequest
import com.nearby.chat.data.remote.dto.DeactivateRequest
import com.nearby.chat.data.remote.dto.LoginRequest
import com.nearby.chat.data.remote.dto.SendCodeRequest
import com.nearby.chat.data.remote.dto.SendNewMobileCodeRequest
import com.nearby.chat.data.remote.dto.UpdateProfileRequest
import com.nearby.chat.data.remote.dto.VerifyOldMobileCodeRequest
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class AuthRepository @Inject constructor(
    private val userApi: UserApi,
    private val tokenStore: TokenStore,
    private val chatLocationStore: ChatLocationStore,
    private val chatRepository: ChatRepository,
    private val deviceRepository: DeviceRepository,
    private val refreshCoordinator: TokenRefreshCoordinator,
) {
    init {
        chatRepository.setTokenRefreshHandler {
            refreshCoordinator.refresh(tokenStore.accessTokenOrNull())
        }
    }

    suspend fun sendLoginCode(mobile: String) {
        userApi.sendLoginCode(SendCodeRequest(mobile)).ensureSuccess()
    }

    suspend fun login(mobile: String, code: String) {
        tokenStore.clear()
        val result = userApi.login(LoginRequest(mobile, code)).unwrap()
        tokenStore.saveSession(result.accessToken, result.refreshToken, result.user.toDomain())
    }

    suspend fun detail(): UserInfo {
        val user = userApi.detail().unwrap().toDomain()
        tokenStore.saveUser(user)
        return user
    }

    suspend fun profile(userId: Long): PublicUserProfile = userApi.profile(userId).unwrap().toDomain()

    suspend fun updateProfile(nickname: String?, avatarMediaId: Long?) {
        userApi.updateProfile(UpdateProfileRequest(nickname, avatarMediaId)).ensureSuccess()
        detail()
    }

    suspend fun sendChangeMobileOldCode() {
        userApi.sendChangeMobileOldCode().ensureSuccess()
    }

    suspend fun verifyChangeMobileOldCode(code: String): String = userApi.verifyChangeMobileOldCode(VerifyOldMobileCodeRequest(code)).unwrap().ticket

    suspend fun sendChangeMobileNewCode(mobile: String, ticket: String) {
        userApi.sendChangeMobileNewCode(SendNewMobileCodeRequest(mobile, ticket)).ensureSuccess()
    }

    suspend fun changeMobile(ticket: String, newMobile: String, newCode: String) {
        userApi.changeMobile(ChangeMobileRequest(ticket, newMobile, newCode)).ensureSuccess()
        detail()
    }

    suspend fun sendDeactivationCode() {
        userApi.sendDeactivationCode().ensureSuccess()
    }

    suspend fun deactivate(code: String) {
        userApi.deactivate(DeactivateRequest(code)).ensureSuccess()
        logout()
    }

    suspend fun blockUser(userId: Long) {
        userApi.block(BlockUserRequest(userId)).ensureSuccess()
    }

    suspend fun unblockUser(userId: Long) {
        userApi.unblock(BlockUserRequest(userId)).ensureSuccess()
    }

    suspend fun blockList(page: Int, pageSize: Int): Pair<Int, List<BlockListItem>> {
        val result = userApi.blocks(page, pageSize).unwrap()
        return result.total to result.list.map { it.toDomain() }
    }

    suspend fun logout() {
        runCatching { deviceRepository.unbind() }
        chatRepository.disconnect()
        tokenStore.clear()
        runCatching { chatLocationStore.clear() }
    }
}
