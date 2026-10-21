package com.nearby.chat.data.remote.api

import com.nearby.chat.core.network.ApiResponse
import com.nearby.chat.data.remote.dto.AdEventsRequest
import com.nearby.chat.data.remote.dto.EmptyData
import retrofit2.http.Body
import retrofit2.http.POST

interface AdApi {
    @POST("api/ad/events")
    suspend fun reportEvents(@Body body: AdEventsRequest): ApiResponse<EmptyData?>
}
