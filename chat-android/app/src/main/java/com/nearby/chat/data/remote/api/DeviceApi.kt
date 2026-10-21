package com.nearby.chat.data.remote.api

import com.nearby.chat.core.network.ApiResponse
import com.nearby.chat.data.remote.dto.DeviceRegisterRequest
import com.nearby.chat.data.remote.dto.EmptyData
import retrofit2.http.Body
import retrofit2.http.POST

interface DeviceApi {
    @POST("api/device/register")
    suspend fun register(@Body body: DeviceRegisterRequest): ApiResponse<EmptyData?>

    @POST("api/device/unbind")
    suspend fun unbind(): ApiResponse<EmptyData?>
}
