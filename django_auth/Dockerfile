FROM python:3.11-slim

WORKDIR /app

# Bağımlılıkları kopyala ve yükle
COPY requirements.txt . 
RUN pip install --no-cache-dir -r requirements.txt

# Proje dosyalarını kopyala
COPY . .

EXPOSE 8000
CMD ["python", "manage.py", "runserver", "0.0.0.0:8000"]