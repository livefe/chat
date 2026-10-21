package com.nearby.chat.data.remote.api

import com.nearby.chat.core.network.ApiResponse
import com.nearby.chat.data.remote.dto.ClearedBeforeResultDto
import com.nearby.chat.data.remote.dto.ConversationListResultDto
import com.nearby.chat.data.remote.dto.ConversationUnreadSummaryDto
import com.nearby.chat.data.remote.dto.EmptyData
import com.nearby.chat.data.remote.dto.PeerIdRequest
import com.nearby.chat.data.remote.dto.PinConversationRequest
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.Query

interface ConversationApi {
    @GET("api/conversation/list")
    suspend fun conversationList(
        @Query("page") page: Int,
        @Query("pageSize") pageSize: Int,
    ): ApiResponse<ConversationListResultDto>

    @GET("api/conversation/unread")
    suspend fun unread(): ApiResponse<ConversationUnreadSummaryDto>

    @POST("api/conversation/read")
    suspend fun markRead(@Body body: PeerIdRequest): ApiResponse<EmptyData?>

    @POST("api/conversation/delete")
    suspend fun deleteConversation(@Body body: PeerIdRequest): ApiResponse<ClearedBeforeResultDto>

    @POST("api/conversation/pin")
    suspend fun pinConversation(@Body body: PinConversationRequest): ApiResponse<EmptyData?>
}
