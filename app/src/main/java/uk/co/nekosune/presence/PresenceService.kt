package uk.co.nekosune.presence

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.Service
import android.content.Intent
import android.os.IBinder
import androidx.core.app.NotificationCompat
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import org.json.JSONObject

class PresenceService : Service() {
    @Volatile private var running = false
    private val http = OkHttpClient.Builder().build()
    private var worker: Thread? = null

    override fun onCreate() {
        super.onCreate()
        createChannel()
        startForeground(7, notification("Waiting for trusted Wi-Fi…"))
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        if (running) return START_STICKY
        running = true
        worker = Thread {
            while (running) {
                val ok = try { heartbeatOnce() } catch (_: Exception) { false }
                val text = if (ok) "Connected to paired PC" else "Paired PC not reachable"
                getSystemService(NotificationManager::class.java).notify(7, notification(text))
                try { Thread.sleep(5000) } catch (_: InterruptedException) { break }
            }
        }.also { it.start() }
        return START_STICKY
    }

    override fun onDestroy() {
        running = false
        worker?.interrupt()
        super.onDestroy()
    }

    override fun onBind(intent: Intent?): IBinder? = null

    private fun heartbeatOnce(): Boolean {
        val host = PairStore.host(this) ?: return false
        val pairId = PairStore.pairId(this) ?: return false
        val computerKey = PairStore.computerKey(this) ?: return false
        if (!NetworkGuard.isTrustedTransport(this) || !NetworkGuard.isPrivateHost(host)) return false

        val challengeReq = Request.Builder().url("http://$host:45873/v1/challenge").get().build()
        val challenge = http.newCall(challengeReq).execute().use { res ->
            if (!res.isSuccessful) return false
            JSONObject(res.body?.string() ?: return false)
        }
        if (challenge.getString("pair_id") != pairId) return false
        if (challenge.getString("computer_public_key") != computerKey) return false
        val nonce = challenge.getString("nonce")
        val serverTs = challenge.getLong("timestamp_ms")
        if (!PresenceProtocol.verifyComputerSignature(computerKey, pairId, nonce, serverTs, challenge.getString("signature_b64"))) return false

        val phoneTs = System.currentTimeMillis()
        val body = JSONObject()
            .put("pair_id", pairId)
            .put("nonce", nonce)
            .put("timestamp_ms", phoneTs)
            .put("signature_b64", PresenceProtocol.sign(pairId, nonce, phoneTs))
            .toString().toRequestBody("application/json".toMediaType())
        val heartbeatReq = Request.Builder().url("http://$host:45873/v1/heartbeat").post(body).build()
        return http.newCall(heartbeatReq).execute().use { it.isSuccessful }
    }

    private fun createChannel() {
        getSystemService(NotificationManager::class.java)
            .createNotificationChannel(NotificationChannel("presence", "NekoPresence", NotificationManager.IMPORTANCE_LOW))
    }

    private fun notification(text: String): Notification = NotificationCompat.Builder(this, "presence")
        .setSmallIcon(android.R.drawable.ic_lock_idle_lock)
        .setContentTitle("NekoPresence Key")
        .setContentText(text)
        .setOngoing(true)
        .build()
}
