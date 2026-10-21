package com.nearby.chat.data.remote.api

import com.nearby.chat.core.network.ApiResponse
import com.nearby.chat.data.remote.dto.CreateRoomRequest
import com.nearby.chat.data.remote.dto.IdResultDto
import com.nearby.chat.data.remote.dto.NearbyBootstrapResultDto
import com.nearby.chat.data.remote.dto.RoomApplicationListResultDto
import com.nearby.chat.data.remote.dto.RoomBootstrapResultDto
import com.nearby.chat.data.remote.dto.RoomDetailDto
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.Path
import retrofit2.http.Query

interface RoomApi {
    @GET("api/room/nearby/bootstrap")
    suspend fun nearbyBootstrap(
        @Query("lat") lat: Double,
        @Query("lng") lng: Double,
        @Query("radius") radius: Int,
        @Query("limit") limit: Int,
        @Query("withMessages") withMessages: Boolean,
    ): ApiResponse<NearbyBootstrapResultDto>

    @GET("api/room/{roomId}")
    suspend fun roomDetail(@Path("roomId") roomId: Long): ApiResponse<RoomDetailDto>

    @GET("api/room/bootstrap")
    suspend fun roomBootstrap(
        @Query("roomId") roomId: Long,
        @Query("limit") limit: Int,
        @Query("withMessages") withMessages: Boolean,
        @Query("lat") lat: Double? = null,
        @Query("lng") lng: Double? = null,
    ): ApiResponse<RoomBootstrapResultDto>

    @POST("api/room/create")
    suspend fun createRoom(@Body body: CreateRoomRequest): ApiResponse<IdResultDto>

    @GET("api/room/applications")
    suspend fun myRoomApplications(
        @Query("page") page: Int,
        @Query("pageSize") pageSize: Int,
    ): ApiResponse<RoomApplicationListResultDto>
}
