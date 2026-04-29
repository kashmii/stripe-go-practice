# stripe-go-practice

Stripe決済フローの実装練習プロジェクト。GoでPaymentIntent作成・Webhook受信・fraudCheckの挿入までを一気通貫で動かすことがゴール。

---

## 目的

- Stripe Payment Intents APIの基本的な使い方を理解する
- WebhookによるイベントドリブンなフローをGoで実装する
- 決済フローの中にfraudCheck関数を挿入し、不正検知の仕組みを体感する
- asukaのJSタグ型不正検知とStripe Radar（APIレイヤー型）の構造的な違いを実装を通じて理解する

---

## 作るもの（最小構成）

1. ユーザーが「購入する」ボタンを押す
2. バックエンド（Go）でPaymentIntentを作成する
3. 決済が成功・失敗したらWebhookで受け取る
4. Webhook受信の直後に `fraudCheck` 関数を挟む

---

## プロジェクト構成

```
stripe-go-practice/
├── main.go
├── handler/
│   ├── payment.go    # PaymentIntent作成
│   └── webhook.go    # Webhook受信
├── fraud/
│   └── check.go      # fraudCheck関数
├── static/
│   └── index.html    # フロントエンド（HTML1枚）
├── README.md
└── CHECKLIST.md
```

---

## 環境変数

```bash
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
```

---

## 使用技術

- Go
- [stripe-go](https://github.com/stripe/stripe-go) v85+
- Stripe CLI（Webhookのローカル転送用）
- Stripe.js + Payment Element（フロントエンド）

---

## テストカード

| カード番号            | 結果                   |
| --------------------- | ---------------------- |
| `4242 4242 4242 4242` | 決済成功               |
| `4000 0000 0000 0002` | 決済失敗（カード拒否） |

---

## 参照

- [Stripe Payment Intents API](https://stripe.com/docs/payments/payment-intents)
- [Stripe Webhooks](https://stripe.com/docs/webhooks)
- [stripe-go](https://github.com/stripe/stripe-go)
- [Stripe CLI](https://stripe.com/docs/stripe-cli)
- 詳細なチェックリストは `CHECKLIST.md` を参照
