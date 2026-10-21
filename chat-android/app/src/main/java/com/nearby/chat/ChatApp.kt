package com.nearby.chat

import android.app.Application
import android.util.Log
import com.nearby.chat.core.lifecycle.AppVisibility
import com.nearby.chat.data.local.DeviceId
import com.nearby.chat.data.local.TokenStore
import com.nearby.chat.data.push.PushChannels
import com.nearby.chat.data.push.PushManager
import com.nearby.chat.data.repo.ChatRepository
import com.nearby.chat.data.repo.DeviceRepository
import com.nearby.chat.di.ApplicationScope
import dagger.hilt.android.HiltAndroidApp
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltAndroidApp
class ChatApp : Application() {

    @Inject lateinit var pushManager: PushManager

    @Inject lateinit var deviceRepository: DeviceRepository

    @Inject lateinit var tokenStore: TokenStore

    @Inject lateinit var chatRepository: ChatRepository

    @Inject lateinit var appVisibility: AppVisibility

    @Inject lateinit var deviceId: DeviceId

    @Inject
    @ApplicationScope
    lateinit var appScope: CoroutineScope

    override fun onCreate() {
        super.onCreate()
        appScope.launch { deviceId.get() }
        appVisibility.install(this)
        appScope.launch(Dispatchers.Main.immediate) {
            appVisibility.foregroundEntered.collect {
                launch {
                    delay(300)
                    if (tokenStore.hasToken.first()) {
                        chatRepository.onAppForeground()
                    }
                }
            }
        }
        PushChannels.ensure(this)
        appScope.launch(Dispatchers.Main.immediate) {
            tokenStore.hasToken
                .distinctUntilChanged()
                .collect { loggedIn ->
                    if (loggedIn) {
                        bootstrapPush()
                        chatRepository.connect()
                    } else {
                        chatRepository.disconnect()
                    }
                }
        }
    }

    private fun bootstrapPush() {
        pushManager.init { token ->
            appScope.launch(Dispatchers.Main.immediate) {
                runCatching { deviceRepository.register(token) }
                    .onFailure { Log.w(TAG, "device register failed", it) }
            }
        }
    }

    companion object {
        private const val TAG = "ChatApp"
    }
}
