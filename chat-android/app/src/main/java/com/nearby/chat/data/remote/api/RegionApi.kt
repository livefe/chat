package com.nearby.chat.data.remote.api

import com.nearby.chat.core.network.ApiResponse
import com.nearby.chat.data.remote.dto.RegionListResponseDto
import retrofit2.http.GET
import retrofit2.http.Query

interface RegionApi {
    @GET("api/region/children")
    suspend fun children(
        @Query("parentCode") parentCode: String = "",
    ): ApiResponse<RegionListResponseDto>

    @GET("api/region/search")
    suspend fun search(
        @Query("keyword") keyword: String,
        @Query("level") level: Int? = null,
        @Query("limit") limit: Int? = null,
    ): ApiResponse<RegionListResponseDto>
}
