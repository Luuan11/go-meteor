let cachedLeaderboard = [];
let lastFetchTime = 0;
const CACHE_DURATION = 30000;
let gameSessionToken = null;
let lastScoreSaveTime = 0;
const MIN_SCORE_INTERVAL = 5000;

const API_URL = 'https://go-meteor.vercel.app/api/leaderboard';
const RECAPTCHA_SITE_KEY = '6LdWFgwsAAAAAAMzR76ilX1OUF56FtKjU2yOlvcG';
const FRONTEND_VERSION = '0.2.0';

console.log('[Frontend Version]', FRONTEND_VERSION);

// Inicializar session token IMEDIATAMENTE (antes do WASM carregar)
(function initSessionImmediately() {
  const timestamp = Date.now();
  const random = Math.random().toString(36).substring(2, 15);
  gameSessionToken = `${timestamp}-${random}`;
  window.gameSessionToken = gameSessionToken;
  console.log('[Session] Token initialized immediately:', gameSessionToken.substring(0, 20) + '...');
})();

async function loadLeaderboard() {
  const now = Date.now();
  
  if (cachedLeaderboard.length > 0 && (now - lastFetchTime) < CACHE_DURATION) {
    return cachedLeaderboard;
  }
  
  try {
    const response = await fetch(API_URL, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    });
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    const data = await response.json();
    
    if (data.leaderboard) {
      cachedLeaderboard = data.leaderboard;
      lastFetchTime = now;
      return cachedLeaderboard;
    }
    return [];
  } catch (error) {
    console.error('Error loading leaderboard:', error);
    return cachedLeaderboard;
  }
}

async function saveScore(playerName, score, signature, timestamp) {
  console.log('[API] saveScore called with:', { playerName, score, hasSignature: !!signature, timestamp });
  
  const trimmedName = playerName ? playerName.trim() : '';
  
  if (!trimmedName || trimmedName.length < 2 || trimmedName.length > 15) {
    console.error('[Validation] Invalid player name: must be 2-15 characters');
    return false;
  }
  
  if (score < 0 || score > 999999) {
    console.error('[Validation] Invalid score range:', score);
    return false;
  }
  
  if (!gameSessionToken) {
    console.error('[Session] No valid session token');
    return false;
  }

  if (!signature) {
    console.error('[Security] Missing HMAC signature');
    return false;
  }

  if (!timestamp) {
    console.error('[Security] Missing timestamp');
    return false;
  }
  
  let recaptchaToken = null;
  if (typeof grecaptcha !== 'undefined' && grecaptcha && grecaptcha.ready) {
    try {
      recaptchaToken = await new Promise((resolve, reject) => {
        const timeout = setTimeout(() => reject(new Error('reCAPTCHA timeout')), 5000);
        grecaptcha.ready(() => {
          grecaptcha.execute(RECAPTCHA_SITE_KEY, { action: 'submit_score' })
            .then(token => { clearTimeout(timeout); resolve(token); })
            .catch(err => { clearTimeout(timeout); reject(err); });
        });
      });
      console.log('[reCAPTCHA] Token generated successfully');
    } catch (error) {
      console.error('[reCAPTCHA] Failed to generate token:', error);
      console.warn('[reCAPTCHA] Continuing without verification');
    }
  } else {
    console.warn('[reCAPTCHA] Not loaded, continuing without verification');
  }
  
  try {
    console.log('[API] Sending POST request to:', API_URL);
    const requestBody = {
      name: trimmedName,
      score: score,
      sessionToken: gameSessionToken,
      timestamp: timestamp,
      signature: signature,
      recaptchaToken: recaptchaToken
    };
    console.log('[API] Request body:', { ...requestBody, signature: signature.substring(0, 16) + '...', recaptchaToken: recaptchaToken ? 'present' : 'null' });
    
    const response = await fetch(API_URL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(requestBody)
    });
    
    if (!response.ok) {
      const error = await response.json();
      console.error('[API] Error response (status ' + response.status + '):', error);
      console.error('[API] Request details:', { name: trimmedName, score, sessionToken: gameSessionToken.substring(0, 16) + '...', timestamp });
      return false;
    }
    
    const result = await response.json();
    console.log('[API] Score saved successfully:', result);
    return true;
    
  } catch (error) {
    console.error('[API] Network error:', error);
    console.error('[API] Error details:', error.message, error.stack);
    return false;
  }
}

function updateLeaderboardUI(leaderboard) {
  const container = document.getElementById('leaderboard-entries');
  if (!container) return;
  
  if (!leaderboard || leaderboard.length === 0) {
    container.innerHTML = '<div class="empty-message">No scores yet. Be the first!</div>';
    return;
  }
  
  container.innerHTML = leaderboard.map((entry, index) => {
    const medal = index === 0 ? '🥇' : index === 1 ? '🥈' : index === 2 ? '🥉' : '';
    const rank = index + 1;
    const rankClass = index < 3 ? `rank-${rank}` : '';
    const displayName = truncateName(entry.name, 20);
    const displayScore = Math.min(entry.score, 999999);
    
    return `
      <div class="leaderboard-entry ${rankClass}">
        <div class="entry-left">
          <span class="entry-rank">${medal || rank}</span>
          <span class="entry-name">${escapeHtml(displayName)}</span>
        </div>
        <span class="entry-score">${displayScore.toLocaleString()}</span>
      </div>
    `;
  }).join('');
}

function truncateName(name, maxLength) {
  if (name.length <= maxLength) return name;
  return name.substring(0, maxLength - 3) + '...';
}

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

window.updateLeaderboard = async function(playerName, score, signature, timestamp) {
  console.log('[Leaderboard] updateLeaderboard called:', { playerName, score, hasSignature: !!signature, timestamp });
  
  if (!gameSessionToken) {
    console.error('[Session] Invalid session - score rejected');
    return false;
  }

  if (!signature || !timestamp) {
    console.error('[Security] Missing signature or timestamp from WASM');
    return false;
  }
  
  const now = Date.now();
  if (now - lastScoreSaveTime < MIN_SCORE_INTERVAL) {
    console.error('[RateLimit] Too many score updates - rate limited');
    return false;
  }
  
  lastScoreSaveTime = now;
  
  const success = await saveScore(playerName, score, signature, timestamp);
  if (success) {
    console.log('[Leaderboard] Score successfully saved to global leaderboard');
    lastFetchTime = 0;
    const leaderboard = await loadLeaderboard();
    updateLeaderboardUI(leaderboard);
    gameSessionToken = null;
    window.gameSessionToken = null;
  } else {
    console.error('[Leaderboard] Failed to save score');
  }
  return success;
};

window.initGameSession = function() {
  const timestamp = Date.now();
  const random = Math.random().toString(36).substring(2, 15);
  gameSessionToken = `${timestamp}-${random}`;
  window.gameSessionToken = gameSessionToken;
  console.log('[Session] Game session initialized');
  return gameSessionToken;
};

window.isTopScore = async function(score) {
  try {
    lastFetchTime = 0;
    const leaderboard = await loadLeaderboard();
    
    console.log('[Leaderboard] Checking score:', score, 'against', leaderboard.length, 'entries');
    
    if (leaderboard.length < 10) {
      console.log('[Leaderboard] Top score: less than 10 entries (has', leaderboard.length + ')');
      return true;
    }
    
    const lowestScore = leaderboard[leaderboard.length - 1].score;
    const isTop = score > lowestScore;
    
    console.log('[Leaderboard] Score check:', score, 'vs lowest:', lowestScore, '- Is top:', isTop);
    return isTop;
  } catch (error) {
    console.error('[Leaderboard] Error checking top score:', error);
    return true;
  }
};

document.addEventListener('DOMContentLoaded', async () => {
  console.log('[Leaderboard] Loading leaderboard UI...');
  
  // Load leaderboard
  const leaderboard = await loadLeaderboard();
  updateLeaderboardUI(leaderboard);
  
  // Auto-refresh every 30 seconds
  setInterval(async () => {
    lastFetchTime = 0;
    const leaderboard = await loadLeaderboard();
    updateLeaderboardUI(leaderboard);
  }, 30000);
});
