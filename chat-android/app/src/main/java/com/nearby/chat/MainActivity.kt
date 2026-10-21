package com.nearby.chat

import android.content.Intent
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.viewModels
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.Surface
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.ui.Modifier
import com.nearby.chat.core.media.VoicePlayer
import com.nearby.chat.core.share.ShareLinks
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.NearbyChatTheme
import com.nearby.chat.core.ui.AppToastHost
import com.nearby.chat.core.ui.LocalToaster
import com.nearby.chat.core.ui.LocalVoicePlayer
import com.nearby.chat.core.ui.ProvideAppTextToolbar
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.feature.main.MainViewModel
import com.nearby.chat.feature.navigation.ChatAppNav
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject

@AndroidEntryPoint
class MainActivity : ComponentActivity() {
    @Inject
    lateinit var toaster: Toaster

    @Inject
    lateinit var voicePlayer: VoicePlayer

    private val mainViewModel: MainViewModel by viewModels()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        mainViewModel.onDeepLink(ShareLinks.parseIntent(intent))
        enableEdgeToEdge()
        setContent {
            CompositionLocalProvider(
                LocalToaster provides toaster,
                LocalVoicePlayer provides voicePlayer,
            ) {
                NearbyChatTheme {
                    Surface(modifier = Modifier.fillMaxSize(), color = AppColors.PageBg) {
                        ProvideAppTextToolbar {
                            Box(modifier = Modifier.fillMaxSize()) {
                                ChatAppNav(mainViewModel = mainViewModel)
                                AppToastHost()
                            }
                        }
                    }
                }
            }
        }
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        setIntent(intent)
        mainViewModel.onDeepLink(ShareLinks.parseIntent(intent))
    }

    override fun onDestroy() {
        if (isFinishing) voicePlayer.stop()
        super.onDestroy()
    }
}
