const admin = require('firebase-admin');
const crypto = require('crypto');

const API_VERSION = '0.2.0';
const MAX_LEADERBOARD_SIZE = 50;
const RATE_LIMIT_WINDOW = 5000;
const RATE_LIMIT_MAP_SIZE = 100;
const TOKEN_CACHE_SIZE = 1000;
const SESSION_TIMEOUT = 30 * 60 * 1000;
const MIN_GAME_TIME = 30;
const MAX_SCORE = 999999;

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
    
    return crypto.timingSafeEqual(
      Buffer.from(signature, 'hex'),
      Buffer.from(expectedSignature, 'hex')
    );
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
  if (isNaN(tokenTimestamp)) {
    return false;
  }
  
  const currentTime = Date.now();

  if (currentTime - tokenTimestamp > SESSION_TIMEOUT) {
    return false;
  }
  
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
  
  if (typeof score !== 'number' || score < 0 || score > MAX_SCORE) {
    return { valid: false, error: 'Invalid score range' };
  }
  
  if (!Number.isInteger(score)) {
    return { valid: false, error: 'Score must be an integer' };
  }
  
  const tokenTimestamp = parseInt(sessionToken.split('-')[0], 10);
  const currentTime = Date.now();
  const gameTimeSeconds = (currentTime - tokenTimestamp) / 1000;
  
  if (gameTimeSeconds < MIN_GAME_TIME) {
    return { valid: false, error: `Game time too short (minimum ${MIN_GAME_TIME} seconds)` };
  }
  
  const averageScorePerSecond = 30;
  const bossBonus = 500;
  const comboMultiplier = 1.5;
  const maxPossibleScore = Math.floor(gameTimeSeconds * averageScorePerSecond * comboMultiplier) + bossBonus;
  
  if (score > maxPossibleScore) {
    return { valid: false, error: `Score too high for game time (max ${maxPossibleScore} in ${Math.floor(gameTimeSeconds)}s)` };
  }
  
  return { valid: true };
}

const rateLimitMap = new Map();
const usedTokens = new Map();

function checkRateLimit(ip) {
  const now = Date.now();
  const lastRequest = rateLimitMap.get(ip);
  
  if (lastRequest && now - lastRequest < RATE_LIMIT_WINDOW) {
    return false;
  }
  
  rateLimitMap.set(ip, now);
  
  if (rateLimitMap.size > RATE_LIMIT_MAP_SIZE) {
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
  
  if (usedTokens.has(sessionToken)) {
    return true;
  }
  
  usedTokens.set(sessionToken, now);
  
  if (usedTokens.size > TOKEN_CACHE_SIZE) {
    const cutoff = now - SESSION_TIMEOUT;
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
  res.setHeader('X-API-Version', API_VERSION);
  
  if (req.method === 'OPTIONS') {
    return res.status(200).end();
  }
  
  const clientIp = req.headers['x-forwarded-for']?.split(',')[0]?.trim() || req.socket.remoteAddress;
  
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
      
      if (!timestamp || typeof timestamp !== 'number') {
        return res.status(400).json({ error: 'Invalid timestamp' });
      }
      
      const currentTime = Date.now();
      const timeDiff = Math.abs(currentTime - timestamp);
      
      if (timeDiff > 60 * 1000) {
        return res.status(403).json({ error: 'Timestamp too old or invalid' });
      }
      
      if (!signature) {
        return res.status(403).json({ error: 'Missing signature' });
      }
      
      if (!verifySignature(name, score, sessionToken, timestamp, signature)) {
        return res.status(403).json({ error: 'Invalid signature' });
      }
      
      if (recaptchaToken) {
        const isValidRecaptcha = await verifyRecaptcha(recaptchaToken);
        if (!isValidRecaptcha) {
          return res.status(403).json({ error: 'reCAPTCHA verification failed' });
        }
      } else {
        console.warn('[reCAPTCHA] No token provided - proceeding without verification');
      }
      
      if (!isValidSession(sessionToken)) {
        return res.status(403).json({ error: 'Invalid session' });
      }
      
      if (isTokenUsed(sessionToken)) {
        return res.status(403).json({ error: 'Session token already used' });
      }
      
      const validation = validateScore(name, score, sessionToken);
      if (!validation.valid) {
        return res.status(400).json({ error: validation.error });
      }
      
      const db = initializeFirebase();
      const newScoreRef = db.ref('leaderboard').push();
      await newScoreRef.set({
        name: name.trim(),
        score: score,
        timestamp: admin.database.ServerValue.TIMESTAMP
      });
      
      try {
        const allScoresSnapshot = await db.ref('leaderboard').orderByChild('score').once('value');
        const allScores = [];
        
        allScoresSnapshot.forEach((child) => {
          allScores.push({
            key: child.key,
            score: child.val().score
          });
        });
        
        allScores.sort((a, b) => b.score - a.score);
        
        if (allScores.length > MAX_LEADERBOARD_SIZE) {
          const toDelete = allScores.slice(MAX_LEADERBOARD_SIZE);
          const deletePromises = toDelete.map(entry => 
            db.ref('leaderboard').child(entry.key).remove()
          );
          await Promise.all(deletePromises);
        }
      } catch (cleanupError) {
        console.error('[Cleanup] Error cleaning leaderboard:', cleanupError);
      }
      
      return res.status(200).json({ 
        success: true,
        message: 'Score saved successfully' 
      });
      
    } else {
      return res.status(405).json({ error: 'Method not allowed' });
    }
    
  } catch (error) {
    console.error('[Error]', error.message);
    return res.status(500).json({ error: 'Internal server error' });
  }
}
