package com.nearby.chat.di

import com.jakewharton.retrofit2.converter.kotlinx.serialization.asConverterFactory
import com.nearby.chat.BuildConfig
import com.nearby.chat.core.network.RefreshHttpClient
import com.nearby.chat.data.network.AuthInterceptor
import com.nearby.chat.data.network.ClientContextInterceptor
import com.nearby.chat.data.network.TokenAuthenticator
import com.nearby.chat.data.remote.api.AdApi
import com.nearby.chat.data.remote.api.AppUpdateApi
import com.nearby.chat.data.remote.api.ConversationApi
import com.nearby.chat.data.remote.api.DeviceApi
import com.nearby.chat.data.remote.api.FeedbackApi
import com.nearby.chat.data.remote.api.MessageApi
import com.nearby.chat.data.remote.api.PostApi
import com.nearby.chat.data.remote.api.RegionApi
import com.nearby.chat.data.remote.api.ReportApi
import com.nearby.chat.data.remote.api.RoomApi
import com.nearby.chat.data.remote.api.SupportApi
import com.nearby.chat.data.remote.api.UploadApi
import com.nearby.chat.data.remote.api.UserApi
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import java.util.concurrent.TimeUnit
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object NetworkModule {

    @Provides
    @Singleton
    fun provideJson(): Json = Json {
        ignoreUnknownKeys = true
        encodeDefaults = true
        explicitNulls = false
    }

    @Provides
    @Singleton
    fun provideOkHttp(
        clientContextInterceptor: ClientContextInterceptor,
        authInterceptor: AuthInterceptor,
        tokenAuthenticator: TokenAuthenticator,
    ): OkHttpClient = OkHttpClient.Builder()
        .connectTimeout(15, TimeUnit.SECONDS)
        .readTimeout(180, TimeUnit.SECONDS)
        .writeTimeout(180, TimeUnit.SECONDS)
        .addInterceptor(clientContextInterceptor)
        .addInterceptor(authInterceptor)
        .authenticator(tokenAuthenticator)
        .addInterceptor(provideLoggingInterceptor())
        .build()

    @Provides
    @Singleton
    @RefreshHttpClient
    fun provideRefreshOkHttp(
        clientContextInterceptor: ClientContextInterceptor,
    ): OkHttpClient = OkHttpClient.Builder()
        .connectTimeout(15, TimeUnit.SECONDS)
        .readTimeout(30, TimeUnit.SECONDS)
        .writeTimeout(30, TimeUnit.SECONDS)
        .addInterceptor(clientContextInterceptor)
        .addInterceptor(provideLoggingInterceptor())
        .build()

    private fun provideLoggingInterceptor(): HttpLoggingInterceptor = HttpLoggingInterceptor().apply {
        level = if (BuildConfig.DEBUG) {
            HttpLoggingInterceptor.Level.HEADERS
        } else {
            HttpLoggingInterceptor.Level.NONE
        }
        redactHeader("Authorization")
    }

    @Provides
    @Singleton
    fun provideRetrofit(client: OkHttpClient, json: Json): Retrofit {
        val contentType = "application/json".toMediaType()
        return Retrofit.Builder()
            .baseUrl(BuildConfig.API_BASE_URL)
            .client(client)
            .addConverterFactory(json.asConverterFactory(contentType))
            .build()
    }

    @Provides
    @Singleton
    fun provideUserApi(retrofit: Retrofit): UserApi = retrofit.create(UserApi::class.java)

    @Provides
    @Singleton
    fun provideRoomApi(retrofit: Retrofit): RoomApi = retrofit.create(RoomApi::class.java)

    @Provides
    @Singleton
    fun provideMessageApi(retrofit: Retrofit): MessageApi = retrofit.create(MessageApi::class.java)

    @Provides
    @Singleton
    fun provideConversationApi(retrofit: Retrofit): ConversationApi = retrofit.create(ConversationApi::class.java)

    @Provides
    @Singleton
    fun providePostApi(retrofit: Retrofit): PostApi = retrofit.create(PostApi::class.java)

    @Provides
    @Singleton
    fun provideAdApi(retrofit: Retrofit): AdApi = retrofit.create(AdApi::class.java)

    @Provides
    @Singleton
    fun provideDeviceApi(retrofit: Retrofit): DeviceApi = retrofit.create(DeviceApi::class.java)

    @Provides
    @Singleton
    fun provideFeedbackApi(retrofit: Retrofit): FeedbackApi = retrofit.create(FeedbackApi::class.java)

    @Provides
    @Singleton
    fun provideSupportApi(retrofit: Retrofit): SupportApi = retrofit.create(SupportApi::class.java)

    @Provides
    @Singleton
    fun provideReportApi(retrofit: Retrofit): ReportApi = retrofit.create(ReportApi::class.java)

    @Provides
    @Singleton
    fun provideRegionApi(retrofit: Retrofit): RegionApi = retrofit.create(RegionApi::class.java)

    @Provides
    @Singleton
    fun provideUploadApi(retrofit: Retrofit): UploadApi = retrofit.create(UploadApi::class.java)

    @Provides
    @Singleton
    fun provideAppUpdateApi(retrofit: Retrofit): AppUpdateApi = retrofit.create(AppUpdateApi::class.java)
}
