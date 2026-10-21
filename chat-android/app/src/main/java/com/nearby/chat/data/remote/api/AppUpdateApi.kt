package com.nearby.chat.data.remote.api

import com.nearby.chat.core.network.ApiResponse
import com.nearby.chat.data.remote.dto.AppUpdateCheckDto
import retrofit2.http.GET

interface AppUpdateApi {
    @GET("api/app/update/check")
    suspend fun check(): ApiResponse<AppUpdateCheckDto>
}
