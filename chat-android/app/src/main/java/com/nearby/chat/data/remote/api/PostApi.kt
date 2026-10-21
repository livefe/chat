package com.nearby.chat.data.remote.api

import com.nearby.chat.core.network.ApiResponse
import com.nearby.chat.data.remote.dto.CreatePostRequest
import com.nearby.chat.data.remote.dto.DiscoverFeedResultDto
import com.nearby.chat.data.remote.dto.EmptyData
import com.nearby.chat.data.remote.dto.IdResultDto
import com.nearby.chat.data.remote.dto.PostCheckInCodeDto
import com.nearby.chat.data.remote.dto.PostCheckInRequest
import com.nearby.chat.data.remote.dto.PostCheckInResultDto
import com.nearby.chat.data.remote.dto.PostItemDto
import com.nearby.chat.data.remote.dto.PostListResultDto
import com.nearby.chat.data.remote.dto.PostSignupListResultDto
import com.nearby.chat.data.remote.dto.PostSignupRequest
import com.nearby.chat.data.remote.dto.SignupStatusResultDto
import com.nearby.chat.data.remote.dto.UpdateSignupRemarkRequest
import com.nearby.chat.data.remote.dto.WithdrawPostRequest
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.Query

interface PostApi {
    @GET("api/post/list")
    suspend fun list(
        @Query("page") page: Int,
        @Query("pageSize") pageSize: Int,
        @Query("latitude") latitude: Double? = null,
        @Query("longitude") longitude: Double? = null,
        @Query("roomId") roomId: Long? = null,
        @Query("category") category: String? = null,
    ): ApiResponse<DiscoverFeedResultDto>

    @GET("api/post/my/list")
    suspend fun myList(
        @Query("page") page: Int,
        @Query("pageSize") pageSize: Int,
        @Query("latitude") latitude: Double? = null,
        @Query("longitude") longitude: Double? = null,
    ): ApiResponse<PostListResultDto>

    @GET("api/post/my/joined/list")
    suspend fun myJoinedList(
        @Query("page") page: Int,
        @Query("pageSize") pageSize: Int,
        @Query("latitude") latitude: Double? = null,
        @Query("longitude") longitude: Double? = null,
    ): ApiResponse<PostListResultDto>

    @GET("api/post/user/list")
    suspend fun userList(
        @Query("userId") userId: Long,
        @Query("page") page: Int,
        @Query("pageSize") pageSize: Int,
        @Query("latitude") latitude: Double? = null,
        @Query("longitude") longitude: Double? = null,
    ): ApiResponse<PostListResultDto>

    @GET("api/post/detail")
    suspend fun detail(@Query("id") id: Long): ApiResponse<PostItemDto>

    @POST("api/post/create")
    suspend fun create(@Body body: CreatePostRequest): ApiResponse<IdResultDto>

    @POST("api/post/update")
    suspend fun update(@Body body: CreatePostRequest): ApiResponse<EmptyData?>

    @POST("api/post/signup")
    suspend fun signup(@Body body: PostSignupRequest): ApiResponse<SignupStatusResultDto>

    @POST("api/post/signup/cancel")
    suspend fun signupCancel(@Body body: PostSignupRequest): ApiResponse<EmptyData?>

    @GET("api/post/signup/check-in-code")
    suspend fun signupCheckInCode(@Query("postId") postId: Long): ApiResponse<PostCheckInCodeDto>

    @POST("api/post/signup/check-in")
    suspend fun signupCheckIn(@Body body: PostCheckInRequest): ApiResponse<PostCheckInResultDto>

    @GET("api/post/signup/list")
    suspend fun signupList(
        @Query("postId") postId: Long,
        @Query("page") page: Int,
        @Query("pageSize") pageSize: Int,
        @Query("status") status: Int? = null,
        @Query("checkIn") checkIn: Int? = null,
    ): ApiResponse<PostSignupListResultDto>

    @POST("api/post/signup/remark")
    suspend fun updateSignupRemark(@Body body: UpdateSignupRemarkRequest): ApiResponse<EmptyData?>

    @POST("api/post/withdraw")
    suspend fun withdraw(@Body body: WithdrawPostRequest): ApiResponse<EmptyData?>

    @POST("api/post/delete")
    suspend fun deleteMine(@Body body: WithdrawPostRequest): ApiResponse<EmptyData?>

    @POST("api/post/my/joined/remove")
    suspend fun removeJoined(@Body body: WithdrawPostRequest): ApiResponse<EmptyData?>
}
