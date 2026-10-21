package com.nearby.chat.data.remote.api

import com.nearby.chat.core.network.ApiResponse
import com.nearby.chat.data.remote.dto.EmptyData
import com.nearby.chat.data.remote.dto.IdRequest
import com.nearby.chat.data.remote.dto.MessageListResultDto
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.Query

interface MessageApi {
    @GET("api/message/room")
    suspend fun roomMessages(
        @Query("roomId") roomId: Long,
        @Query("beforeId") beforeId: Long,
        @Query("limit") limit: Int,
        @Query("lat") lat: Double? = null,
        @Query("lng") lng: Double? = null,
    ): ApiResponse<MessageListResultDto>

    @GET("api/message/private")
    suspend fun privateMessages(
        @Query("peerId") peerId: Long,
        @Query("beforeId") beforeId: Long,
        @Query("limit") limit: Int,
    ): ApiResponse<MessageListResultDto>

    @GET("api/message/post")
    suspend fun postMessages(
        @Query("postId") postId: Long,
        @Query("beforeId") beforeId: Long,
        @Query("limit") limit: Int,
    ): ApiResponse<MessageListResultDto>

    @GET("api/message/nearby")
    suspend fun nearbyMessages(
        @Query("lat") lat: Double,
        @Query("lng") lng: Double,
        @Query("beforeId") beforeId: Long,
        @Query("radius") radius: Int,
        @Query("limit") limit: Int,
    ): ApiResponse<MessageListResultDto>

    @POST("api/message/recall")
    suspend fun recallMessage(@Body body: IdRequest): ApiResponse<EmptyData?>
}
