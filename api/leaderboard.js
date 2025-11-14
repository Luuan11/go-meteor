const admin = require('firebase-admin');
const crypto = require('crypto');

const API_VERSION = '0.2.0';
console.log('[API Version]', API_VERSION);

let firebaseApp;

function getSecretKey() {
  const encrypted = Buffer.from([
    0x67, 0x6f, 0x2d, 0x6d, 0x65, 0x74, 0x65, 0x6f, 0x72, 0x2d,
    0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x2d, 0x32, 0x30, 0x32,
    0x35, 0x2d, 0x73, 0x65, 0x63, 0x75, 0x72, 0x65
  ]);
  const key = Buffer.alloc(encrypted.length);
  for (let i = 0; i < encrypted.length; i++) {
    key[i] = encrypted[i] ^ 0xAA;
  }
  return key;
}

function verifySignature(name, score, sessionToken, timestamp, signature) {
  try {
    const message = `${name}|${score}|${sessionToken}|${timestamp}`;
    const hmac = crypto.createHmac('sha256', getSecretKey());
    hmac.update(message);
    const expectedSignature = hmac.digest('hex');
    
    const isValid = crypto.timingSafeEqual(
      Buffer.from(signature, 'hex'),
      Buffer.from(expectedSignature, 'hex')
    );
    
    if (isValid) {
      console.log('[Security] HMAC signature verified successfully');
    } else {
      console.log('[Security] HMAC signature verification failed');
      console.log('[Security] Expected:', expectedSignature.substring(0, 16) + '...');
      console.log('[Security] Received:', signature.substring(0, 16) + '...');
    }
    
    return isValid;
  } catch (error) {
    console.error('[Security] Signature verification error:', error);
    return false;
  }
}

async function verifyRecaptcha(token) {
  const secretKey = process.env.RECAPTCHA_SECRET_KEY;
  
  if (!secretKey) {
    console.error('RECAPTCHA_SECRET_KEY not configured');
    return false;
  }
  
  try {
    const response = await fetch('https://www.google.com/recaptcha/api/siteverify', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: `secret=${secretKey}&response=${token}`
    });
    
    const data = await response.json();
    
    if (data.success && data.score >= 0.5) {
      return true;
    }
    
    console.log('reCAPTCHA verification failed:', data);
    return false;
  } catch (error) {
    console.error('reCAPTCHA verification error:', error);
    return false;
  }
}

function initializeFirebase() {
  if (!firebaseApp) {
    firebaseApp = admin.initializeApp({
      credential: admin.credential.cert({
        projectId: process.env.FIREBASE_PROJECT_ID,
        clientEmail: process.env.FIREBASE_CLIENT_EMAIL,
        privateKey: process.env.FIREBASE_PRIVATE_KEY.replace(/\\n/g, '\n'),
      }),
      databaseURL: process.env.FIREBASE_DATABASE_URL,
    });
  }
  return admin.database();
}

function isValidSession(sessionToken) {
  if (!sessionToken || typeof sessionToken !== 'string') {
    return false;
  }
  
  const parts = sessionToken.split('-');
  if (parts.length !== 2) {
    return false;
  }
  
  const tokenTimestamp = parseInt(parts[0], 10);
  const currentTime = Date.now();

  // Token deve ter no máximo 30 minutos (tempo razoável de uma partida)
  if (currentTime - tokenTimestamp > 30 * 60 * 1000) {
    return false;
  }
  
  // Token não pode ser do futuro (tolerância de 1 minuto para diferenças de clock)
  if (tokenTimestamp - currentTime > 60 * 1000) {
    return false;
  }
  
  return true;
}

function validateScore(name, score, sessionToken) {
  const trimmedName = typeof name === 'string' ? name.trim() : '';
  
  if (!trimmedName || trimmedName.length < 2 || trimmedName.length > 15) {
    return { valid: false, error: 'Invalid name length (2-15 characters)' };
  }
  
  if (!/^[a-zA-Z0-9 ]+$/.test(trimmedName)) {
    return { valid: false, error: 'Invalid name characters' };
  }
  
  if (typeof score !== 'number' || score < 0 || score > 999999) {
    return { valid: false, error: 'Invalid score range' };
  }
  
  const tokenTimestamp = parseInt(sessionToken.split('-')[0], 10);
  const currentTime = Date.now();
  const gameTimeSeconds = (currentTime - tokenTimestamp) / 1000;
  
  if (gameTimeSeconds < 30) {
    return { valid: false, error: 'Game time too short (minimum 30 seconds)' };
  }
  
  const maxScorePerSecond = 10;
  const maxPossibleScore = Math.floor(gameTimeSeconds * maxScorePerSecond);
  
  if (score > maxPossibleScore) {
    return { valid: false, error: `Score too high for game time (max ${maxPossibleScore} in ${Math.floor(gameTimeSeconds)}s)` };
  }
  
  return { valid: true };
}

const rateLimitMap = new Map();
const usedTokens = new Map(); // Blacklist de tokens já usados

function checkRateLimit(ip) {
  const now = Date.now();
  const lastRequest = rateLimitMap.get(ip);
  
  if (lastRequest && now - lastRequest < 5000) {
    return false;
  }
  
  rateLimitMap.set(ip, now);
  
  if (rateLimitMap.size > 100) {
    const cutoff = now - 60000;
    for (const [key, value] of rateLimitMap.entries()) {
      if (value < cutoff) {
        rateLimitMap.delete(key);
      }
    }
  }
  
  return true;
}

function isTokenUsed(sessionToken) {
  const now = Date.now();
  
  // Verifica se o token já foi usado
  if (usedTokens.has(sessionToken)) {
    return true;
  }
  
  // Marca o token como usado
  usedTokens.set(sessionToken, now);
  
  // Limpa tokens antigos (mais de 15 minutos)
  if (usedTokens.size > 1000) {
    const cutoff = now - 15 * 60 * 1000;
    for (const [key, value] of usedTokens.entries()) {
      if (value < cutoff) {
        usedTokens.delete(key);
      }
    }
  }
  
  return false;
}

export default async function handler(req, res) {

  res.setHeader('Access-Control-Allow-Origin', 'https://luuan11.github.io');
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type');
  
  if (req.method === 'OPTIONS') {
    return res.status(200).end();
  }
  
  const clientIp = req.headers['x-forwarded-for'] || req.socket.remoteAddress;
  
  try {
    if (req.method === 'GET') {

      const db = initializeFirebase();
      const snapshot = await db.ref('leaderboard').orderByChild('score').limitToLast(10).once('value');
      
      const leaderboard = [];
      snapshot.forEach((child) => {
        leaderboard.push({
          id: child.key,
          ...child.val()
        });
      });
      
      leaderboard.sort((a, b) => b.score - a.score);
      
      return res.status(200).json({ leaderboard });
      
    } else if (req.method === 'POST') {
      
      if (!checkRateLimit(clientIp)) {
        return res.status(429).json({ error: 'Too many requests. Please wait.' });
      }
      
      const { name, score, sessionToken, timestamp, signature, recaptchaToken } = req.body;
      
      console.log('[API] POST request received from IP:', clientIp);
      console.log('[API] Request data:', { name, score, hasSignature: !!signature, timestamp, hasRecaptcha: !!recaptchaToken });
      
      if (!signature) {
        console.log('[Security] Missing HMAC signature');
        return res.status(403).json({ error: 'Missing signature' });
      }
      
      if (!verifySignature(name, score, sessionToken, timestamp, signature)) {
        console.log('[Security] Invalid HMAC signature - possible tampering detected');
        return res.status(403).json({ error: 'Invalid signature' });
      }
      
      if (recaptchaToken) {
        const isValidRecaptcha = await verifyRecaptcha(recaptchaToken);
        if (!isValidRecaptcha) {
          console.log('[reCAPTCHA] Verification failed');
          return res.status(403).json({ error: 'reCAPTCHA verification failed' });
        }
        console.log('[reCAPTCHA] Verified successfully');
      } else {
        console.warn('[reCAPTCHA] No token provided - proceeding without verification');
      }
      
      if (!isValidSession(sessionToken)) {
        console.log('[Session] Invalid session:', { sessionToken: sessionToken.substring(0, 16) + '...', timestamp });
        return res.status(403).json({ error: 'Invalid session' });
      }
      console.log('[Session] Session token validated');
      
      if (isTokenUsed(sessionToken)) {
        console.log('[Session] Token already used:', sessionToken.substring(0, 16) + '...');
        return res.status(403).json({ error: 'Session token already used' });
      }
      console.log('[Session] Token marked as used');
      
      const validation = validateScore(name, score, sessionToken);
      if (!validation.valid) {
        console.log('[Validation] Score validation failed:', validation.error, { name, score });
        return res.status(400).json({ error: validation.error });
      }
      console.log('[Validation] Score validated successfully');
      
      const db = initializeFirebase();
      const newScoreRef = db.ref('leaderboard').push();
      await newScoreRef.set({
        name: name.trim(),
        score: score,
        timestamp: admin.database.ServerValue.TIMESTAMP
      });
      
      console.log('[Database] Score saved successfully:', { name: name.trim(), score, ip: clientIp });
      
      return res.status(200).json({ 
        success: true,
        message: 'Score saved successfully' 
      });
      
    } else {
      return res.status(405).json({ error: 'Method not allowed' });
    }
    
  } catch (error) {
    console.error('Error:', error);
    return res.status(500).json({ error: 'Internal server error' });
  }
}
