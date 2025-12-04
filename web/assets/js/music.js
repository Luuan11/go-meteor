(function() {
    const bgMusic = document.getElementById('bgMusic');
    if (!bgMusic) return;
    
    bgMusic.volume = 0.3;
    
    const playPromise = bgMusic.play();
    
    if (playPromise !== undefined) {
        playPromise.catch(() => {
            const startMusic = () => {
                bgMusic.play();
                document.removeEventListener('click', startMusic);
                document.removeEventListener('keydown', startMusic);
                document.removeEventListener('touchstart', startMusic);
            };
            
            document.addEventListener('click', startMusic);
            document.addEventListener('keydown', startMusic);
            document.addEventListener('touchstart', startMusic);
        });
    }
})();
