package com.nearby.chat.data.remote.api

import com.nearby.chat.core.network.ApiResponse
import com.nearby.chat.data.remote.dto.UploadMediaDto
import okhttp3.MultipartBody
import retrofit2.http.Multipart
import retrofit2.http.POST
import retrofit2.http.Part

interface UploadApi {
    @Multipart
    @POST("api/upload/image")
    suspend fun uploadImage(@Part file: MultipartBody.Part): ApiResponse<UploadMediaDto>

    @Multipart
    @POST("api/upload/video")
    suspend fun uploadVideo(@Part file: MultipartBody.Part): ApiResponse<UploadMediaDto>

    @Multipart
    @POST("api/upload/audio")
    suspend fun uploadAudio(@Part file: MultipartBody.Part): ApiResponse<UploadMediaDto>

    @Multipart
    @POST("api/upload/file")
    suspend fun uploadFile(@Part file: MultipartBody.Part): ApiResponse<UploadMediaDto>
}
