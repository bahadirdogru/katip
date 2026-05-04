# Katip - Editöryal Verimlilik ve Akıllı Çalışma Planlaması (Kesinleşmiş Mimari)

## 1. Sistemin Ana Hedefi
Katip; yayınevi çalışanlarının ve yazarların **işine (metne) en yüksek düzeyde odaklanmasını** sağlamak üzere tasarlanmış, bireysel verimliliği maksimize eden, şifreli takım karmaşasından uzak bir masaüstü (Windows, macOS, Linux) metin editörüdür. 

Tüm gücünü yerel cihazdaki AI'dan (LLM) alır; internete hiçbir veri göndermez. Sektörel olarak en sıkıntı çekilen asenkron işbirliği ve geri bildirim sürecini; tüm dosyaların, eklentilerin, geçmişin ve profillerin tek bir **e-posta paketi (.kitap)** içine hapsedilmesiyle çözer.

---

## 2. Kesinleşmiş Geliştirme Fazları (Epic Planları)

Kullanıcı gereksinimleri sonrası netleşmiş taslak özellik haritası aşağıdaki gibidir:

### Faz 1: "Her Şey Dahil" Offline Paket (`.kitap`)
En güçlü yenilik, e-posta akışında hiçbir şeyin yolda "kaybolmamasıdır".
- **Genişletilmiş Paket Yapısı:** İçerisinde Markdown formatında metin dosyalarını, değişim geçmişini, yazar/editör yorumlarını, gömülü projesel meta verileri (yazar, çevirmen, kelime numarası) ve aynı zamanda metne eklenen referans Resimleri/Medyaları, hatta özel Fontları dahi sıkıştırılmış tek bir dosya olarak tutan format.
- **İçerik Bölümleme (Chaptering):** Projenin bölümlere ayrılıp sol panel vasıtasıyla kolayca dolaşılması.
- **Proje Özeti (Master Summary):** `.kitap` dosyasının kalbinde, karakterlerin (göz rengi, fizyolojik durumu) ve olay akışının basitçe tanımlar halinde tutulduğu entegre bir özet panelinin bulunması ve sürekli güncellenmesi.

### Faz 2: İzleme, Temiz UI ve Editöryal Odak
Merge Conflict (Çatışma Yönetimi) veya hiyerarşik yapılandırma şimdilik dahil düşünülmeyecektir.
- **Cümle/Satır Bazlı Hassas Takip (Git Mantığı):** Word'ün gürültülü "Değişikliği İzlendir" özelliği yerine; çok daha net, temiz ve satır satır "Ne eklendi / Ne çıkarıldı" gösteren geçmiş sistemi. Versiyonlama sadece "Kullanıcı Adı, Tarih ve Saat" evrensel mantığı ile tutulacaktır.
- **Yorumlar (Comments) ve Etiketlemeler:** Yayıncıların asenkron iletişimi için PDF çıktısına gitmeyecek içsel notlaşma (yorum baloncukları) sistemi.
- **Melez ve Eğitici Arayüz (Word/Notion):** Yukarıda mutlaka erişilebilir ama şık bir Araç Çubuğu (Toolbar) barındırılacak. Bu araç çubuğundaki düğmelerin üzerine imleç getirilince çıkan (Tooltip) ipuçları ile editörler Markdown ve klavye kısayollarını kullanmaya teşvik edilecek.
- **Çalışma Ortamı Ergonomisi:** Kesintisiz Gece ve Gündüz modları, metni rahat okumak için ölçeklendirme (Sayfa Büyütme/Zoom) özellikleri eklenecektir.

### Faz 3: LLM'in Yayınevi Beynine (Kurallara) Dönüştürülmesi
- **`.tarz` Uzantılı Profil İçe/Dışa Aktarımı:** Katip, yayınevinin veya yazarın standart jargonuna riayet edilmesi, özel kelimelerin düzeltilmesi ve stil standartlarının korunması için oluşturulmuş özel AI Şablon Profillerini `.tarz` uzantılı bir dosya olarak (flash bellek gibi) üretebilecek ve dışarıdan başka cihazlara import edebilecektir.
- **Kitap Özeti Üzerinden Tutarlık Denetleyicisi (RAG):** AI uzun, karmaşık tam metni taramak yerine; Faz 1'de belirtilen projenin içine gömülü "Olay ve Karakter Özeti" referans kartını kullanarak okuduğu bir sahnede, özetle mantık çerçevesinde uyuşmayan yerleri (karakter özellikleri veya zaman sırası hataları) yakalayıp editöre anlık bildirecektir. Bu sayede RAM kullanımı minimumda tutulup anında reaksiyon alınacaktır.
- **Çok Yönlü İlk Okuma / Seri Tarama:** Tek tıkla o andaki çalıştığınız bölüm, anlatım bozukluklarına, yanlış noktalama işaretlerine ve seçilen aktif `.tarz` profiline göre tamamen taranıp bulgular bir "Öneri Kartları" listesi olarak kullanıcının onayına sunulacaktır.

---

*(Not: IDML, DOCX, EPUB dışa/içe aktarımları bu versiyon paketlemesinde hedeflenmemiş, sadece Markdown ekseninde kalınması planlanmıştır.)*
