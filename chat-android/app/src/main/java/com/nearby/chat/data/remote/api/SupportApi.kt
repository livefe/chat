package com.nearby.chat.data.remote.api

import com.nearby.chat.core.network.ApiResponse
import com.nearby.chat.data.remote.dto.EmptyData
import com.nearby.chat.data.remote.dto.SupportMessageDto
import com.nearby.chat.data.remote.dto.SupportMessageListResultDto
import com.nearby.chat.data.remote.dto.SupportOpenRequest
import com.nearby.chat.data.remote.dto.SupportProfileDto
import com.nearby.chat.data.remote.dto.SupportReadRequest
import com.nearby.chat.data.remote.dto.SupportSendRequest
import com.nearby.chat.data.remote.dto.SupportSessionDto
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.Query

interface SupportApi {
    @GET("api/support/profile")
    suspend fun profile(): ApiResponse<SupportProfileDto>

    @GET("api/support/session/current")
    suspend fun current(): ApiResponse<SupportSessionDto?>

    @POST("api/support/session/open")
    suspend fun open(@Body body: SupportOpenRequest): ApiResponse<SupportSessionDto>

    @GET("api/support/messages")
    suspend fun messages(
        @Query("conversationId") conversationId: Long,
        @Query("beforeId") beforeId: Long? = null,
        @Query("limit") limit: Int,
    ): ApiResponse<SupportMessageListResultDto>

    @POST("api/support/messages")
    suspend fun send(@Body body: SupportSendRequest): ApiResponse<SupportMessageDto>

    @POST("api/support/read")
    suspend fun read(@Body body: SupportReadRequest): ApiResponse<EmptyData?>
}
