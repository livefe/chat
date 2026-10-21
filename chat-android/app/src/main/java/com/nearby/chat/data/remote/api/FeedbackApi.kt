package com.nearby.chat.data.remote.api

import com.nearby.chat.core.network.ApiResponse
import com.nearby.chat.data.remote.dto.EmptyData
import com.nearby.chat.data.remote.dto.FeedbackItemDto
import com.nearby.chat.data.remote.dto.FeedbackListResultDto
import com.nearby.chat.data.remote.dto.SubmitFeedbackRequest
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.Query

interface FeedbackApi {
    @POST("api/feedback/submit")
    suspend fun submit(@Body body: SubmitFeedbackRequest): ApiResponse<EmptyData?>

    @GET("api/feedback/my/list")
    suspend fun myList(
        @Query("page") page: Int,
        @Query("pageSize") pageSize: Int,
    ): ApiResponse<FeedbackListResultDto>

    @GET("api/feedback/my/detail")
    suspend fun myDetail(@Query("id") id: Long): ApiResponse<FeedbackItemDto>
}
