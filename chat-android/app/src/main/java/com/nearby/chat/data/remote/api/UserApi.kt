package com.nearby.chat.data.remote.api

import com.nearby.chat.core.network.ApiResponse
import com.nearby.chat.data.remote.dto.BlockListResultDto
import com.nearby.chat.data.remote.dto.BlockUserRequest
import com.nearby.chat.data.remote.dto.ChangeMobileRequest
import com.nearby.chat.data.remote.dto.DeactivateRequest
import com.nearby.chat.data.remote.dto.EmptyData
import com.nearby.chat.data.remote.dto.LoginRequest
import com.nearby.chat.data.remote.dto.LoginResultDto
import com.nearby.chat.data.remote.dto.PublicUserProfileDto
import com.nearby.chat.data.remote.dto.SendCodeRequest
import com.nearby.chat.data.remote.dto.SendNewMobileCodeRequest
import com.nearby.chat.data.remote.dto.UpdateProfileRequest
import com.nearby.chat.data.remote.dto.UserInfoDto
import com.nearby.chat.data.remote.dto.VerifyOldMobileCodeRequest
import com.nearby.chat.data.remote.dto.VerifyOldMobileCodeResultDto
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.Query

interface UserApi {
    @POST("api/auth/login")
    suspend fun login(@Body body: LoginRequest): ApiResponse<LoginResultDto>

    @GET("api/user/me")
    suspend fun detail(): ApiResponse<UserInfoDto>

    @GET("api/user/profile")
    suspend fun profile(@Query("id") id: Long): ApiResponse<PublicUserProfileDto>

    @POST("api/user/profile/update")
    suspend fun updateProfile(@Body body: UpdateProfileRequest): ApiResponse<EmptyData?>

    @POST("api/auth/sms/code")
    suspend fun sendLoginCode(@Body body: SendCodeRequest): ApiResponse<EmptyData?>

    @POST("api/user/mobile/send/old-code")
    suspend fun sendChangeMobileOldCode(): ApiResponse<EmptyData?>

    @POST("api/user/mobile/verify/old-code")
    suspend fun verifyChangeMobileOldCode(@Body body: VerifyOldMobileCodeRequest): ApiResponse<VerifyOldMobileCodeResultDto>

    @POST("api/user/mobile/send/new-code")
    suspend fun sendChangeMobileNewCode(@Body body: SendNewMobileCodeRequest): ApiResponse<EmptyData?>

    @POST("api/user/mobile/change")
    suspend fun changeMobile(@Body body: ChangeMobileRequest): ApiResponse<EmptyData?>

    @POST("api/user/deactivation/send/code")
    suspend fun sendDeactivationCode(): ApiResponse<EmptyData?>

    @POST("api/user/deactivate")
    suspend fun deactivate(@Body body: DeactivateRequest): ApiResponse<EmptyData?>

    @POST("api/user/block")
    suspend fun block(@Body body: BlockUserRequest): ApiResponse<EmptyData?>

    @POST("api/user/unblock")
    suspend fun unblock(@Body body: BlockUserRequest): ApiResponse<EmptyData?>

    @GET("api/user/blocks")
    suspend fun blocks(
        @Query("page") page: Int,
        @Query("pageSize") pageSize: Int,
    ): ApiResponse<BlockListResultDto>
}
