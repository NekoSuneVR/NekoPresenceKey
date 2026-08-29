package uk.co.nekosune.presence

import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import android.util.Base64
import java.security.KeyFactory
import java.security.KeyPairGenerator
import java.security.KeyStore
import java.security.Signature
import java.security.spec.X509EncodedKeySpec

object PresenceProtocol {
    private const val ALIAS = "NekoPresencePhoneKey"

    fun ensureKey() {
        val ks = KeyStore.getInstance("AndroidKeyStore").apply { load(null) }
        if (ks.containsAlias(ALIAS)) return
        val generator = KeyPairGenerator.getInstance(KeyProperties.KEY_ALGORITHM_ED25519, "AndroidKeyStore")
        generator.initialize(KeyGenParameterSpec.Builder(ALIAS, KeyProperties.PURPOSE_SIGN or KeyProperties.PURPOSE_VERIFY).build())
        generator.generateKeyPair()
    }

    fun publicKeyRawHex(): String {
        ensureKey()
        val ks = KeyStore.getInstance("AndroidKeyStore").apply { load(null) }
        val encoded = ks.getCertificate(ALIAS).publicKey.encoded
        val raw = encoded.copyOfRange(encoded.size - 32, encoded.size)
        return raw.joinToString("") { "%02x".format(it) }
    }

    fun sign(pairId: String, nonce: String, timestampMs: Long): String {
        ensureKey()
        val ks = KeyStore.getInstance("AndroidKeyStore").apply { load(null) }
        val privateKey = ks.getKey(ALIAS, null)
        val sig = Signature.getInstance("Ed25519")
        sig.initSign(privateKey as java.security.PrivateKey)
        sig.update("$pairId|$nonce|$timestampMs".toByteArray())
        return Base64.encodeToString(sig.sign(), Base64.NO_WRAP)
    }

    fun verifyComputerSignature(computerRawHex: String, pairId: String, nonce: String, timestampMs: Long, signatureB64: String): Boolean {
        return try {
            val raw = computerRawHex.chunked(2).map { it.toInt(16).toByte() }.toByteArray()
            val prefix = byteArrayOf(0x30, 0x2a, 0x30, 0x05, 0x06, 0x03, 0x2b, 0x65, 0x70, 0x03, 0x21, 0x00)
            val pub = KeyFactory.getInstance("Ed25519").generatePublic(X509EncodedKeySpec(prefix + raw))
            val sig = Signature.getInstance("Ed25519")
            sig.initVerify(pub)
            sig.update("$pairId|$nonce|$timestampMs".toByteArray())
            sig.verify(Base64.decode(signatureB64, Base64.DEFAULT))
        } catch (_: Exception) {
            false
        }
    }
}
