package com.nearby.chat.data.remote.api

import com.nearby.chat.core.network.ApiResponse
import com.nearby.chat.data.remote.dto.EmptyData
import com.nearby.chat.data.remote.dto.SubmitReportRequest
import retrofit2.http.Body
import retrofit2.http.POST

interface ReportApi {
    @POST("api/report/submit")
    suspend fun submit(@Body body: SubmitReportRequest): ApiResponse<EmptyData?>
}
