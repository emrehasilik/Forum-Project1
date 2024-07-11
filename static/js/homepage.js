// JavaScript kodunu ekleyin
document.addEventListener('DOMContentLoaded', function() {
    // Yorum popup elemanlarını alın
    var yorumPopup = document.getElementById('yorumPopup');
    var yorumPopupClose = document.querySelector('.yorum-popup-close');
    var commentButtons = document.querySelectorAll('.commentButton');
    var yorumEkleButton = document.getElementById('yorumEkleButton');
    var yorumInput = document.getElementById('yorumInput');
    var yorumListe = document.querySelector('.yorum-liste');

    // Yorum yap butonuna tıklanınca popup'ı göster
    commentButtons.forEach(function(button) {
        button.addEventListener('click', function() {
            yorumPopup.style.display = 'block';
        });
    });

    // Kapatma butonuna tıklanınca popup'ı gizle
    yorumPopupClose.addEventListener('click', function() {
        yorumPopup.style.display = 'none';
    });

    // Popup dışında bir yere tıklanınca popup'ı gizle
    window.addEventListener('click', function(event) {
        if (event.target == yorumPopup) {
            yorumPopup.style.display = 'none';
        }
    });

    // Yorum ekle butonuna tıklanınca yeni yorumu ekle
    yorumEkleButton.addEventListener('click', function() {
        var yorumText = yorumInput.value.trim();
        if (yorumText) {
            var yeniYorum = document.createElement('div');
            yeniYorum.className = 'yorum';
            yeniYorum.innerHTML = '<p><strong>Benim Yorumum:</strong> ' + yorumText + '</p><div class="yorum-footer"><button class="yorum-begen-button"><span class="material-icons">thumb_up</span> 0</button><button class="yorum-begenme-button"><span class="material-icons">thumb_down</span> 0</button></div>';
            yorumListe.appendChild(yeniYorum);
            yorumInput.value = '';
        }
    });

    // Dinamik olarak eklenen yorumların beğen ve beğenme butonlarını yönetme
    document.addEventListener('click', function(event) {
        if (event.target.matches('.yorum-begen-button')) {
            var begenButton = event.target.closest('.yorum-begen-button');
            var countSpan = begenButton.querySelector('span');
            var count = parseInt(countSpan.textContent);
            countSpan.textContent = count + 1;
        } else if (event.target.matches('.yorum-begenme-button')) {
            var begenmeButton = event.target.closest('.yorum-begenme-button');
            var countSpan = begenmeButton.querySelector('span');
            var count = parseInt(countSpan.textContent);
            countSpan.textContent = count + 1;
        }
    });
});
