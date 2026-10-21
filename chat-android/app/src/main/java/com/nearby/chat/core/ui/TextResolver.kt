package com.nearby.chat.core.ui

import android.content.Context
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TextResolver @Inject constructor(
    @ApplicationContext private val context: Context,
) {
    fun resolve(text: UiText): String = text.resolve(context)
}
