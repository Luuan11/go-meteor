const admin = require('firebase-admin');

let firebaseApp;

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

function validateScore(name, score) {
  if (typeof name !== 'string' || name.length < 2 || name.length > 20) {
    return { valid: false, error: 'Invalid name length' };
  }
  
  if (!/^[a-zA-Z0-9 ]+$/.test(name)) {
    return { valid: false, error: 'Invalid name characters' };
  }
  
  if (typeof score !== 'number' || score < 0 || score > 999999) {
    return { valid: false, error: 'Invalid score range' };
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
      
      const { name, score, sessionToken, timestamp, recaptchaToken } = req.body;
      
      if (recaptchaToken) {
        const isValidRecaptcha = await verifyRecaptcha(recaptchaToken);
        if (!isValidRecaptcha) {
          return res.status(403).json({ error: 'reCAPTCHA verification failed' });
        }
      }
      
      if (!isValidSession(sessionToken)) {
        console.log('Invalid session:', { sessionToken, timestamp });
        return res.status(403).json({ error: 'Invalid session' });
      }
      
      // Verifica se o token já foi usado (proteção contra replay attack)
      if (isTokenUsed(sessionToken)) {
        console.log('Token already used:', sessionToken);
        return res.status(403).json({ error: 'Session token already used' });
      }
      
      const validation = validateScore(name, score);
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
      
      console.log('Score saved:', { name, score, ip: clientIp });
      
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
