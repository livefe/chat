package com.nearby.chat.data.push

import dagger.hilt.EntryPoint
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent

@EntryPoint
@InstallIn(SingletonComponent::class)
interface PushEntryPoint {
    fun pushTokenStore(): PushTokenStore
}
