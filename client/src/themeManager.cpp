#include "themeManager.h"

#include <QApplication>
#include <QFile>
#include <QPalette>
#include <QSettings>
#include <QStyleFactory>

ThemeManager::Theme ThemeManager::theme() {
    QSettings s;
    return themeFromString(s.value(settingsKey(), QStringLiteral("dark")).toString());
}

void ThemeManager::setTheme(ThemeManager::Theme theme) {
    QSettings s;
    s.setValue(settingsKey(), themeToString(theme));

    if (auto* app = qobject_cast<QApplication*>(QCoreApplication::instance())) {
        applyTo(*app);
    }
}

static QPalette paletteForTheme(ThemeManager::Theme theme) {
    QPalette palette;

    if (theme == ThemeManager::Theme::Dark) {
        palette.setColor(QPalette::Window, QColor("#0b1220"));
        palette.setColor(QPalette::WindowText, QColor("#e5e7eb"));
        palette.setColor(QPalette::Base, QColor("#0f172a"));
        palette.setColor(QPalette::AlternateBase, QColor("#111827"));
        palette.setColor(QPalette::Text, QColor("#e5e7eb"));
        palette.setColor(QPalette::Button, QColor("#111827"));
        palette.setColor(QPalette::ButtonText, QColor("#e5e7eb"));
        palette.setColor(QPalette::Highlight, QColor("#1d4ed8"));
        palette.setColor(QPalette::HighlightedText, QColor("#ffffff"));
        palette.setColor(QPalette::ToolTipBase, QColor("#111827"));
        palette.setColor(QPalette::ToolTipText, QColor("#e5e7eb"));
        return palette;
    }

    palette.setColor(QPalette::Window, QColor("#f7f8fa"));
    palette.setColor(QPalette::WindowText, QColor("#111827"));
    palette.setColor(QPalette::Base, QColor("#ffffff"));
    palette.setColor(QPalette::AlternateBase, QColor("#f9fafb"));
    palette.setColor(QPalette::Text, QColor("#111827"));
    palette.setColor(QPalette::Button, QColor("#ffffff"));
    palette.setColor(QPalette::ButtonText, QColor("#111827"));
    palette.setColor(QPalette::Highlight, QColor("#dbeafe"));
    palette.setColor(QPalette::HighlightedText, QColor("#111827"));
    return palette;
}

static QString qssResourceForTheme(ThemeManager::Theme theme) {
    return theme == ThemeManager::Theme::Dark ? QStringLiteral(":/styles/app_dark.qss") : QStringLiteral(":/styles/app.qss");
}

void ThemeManager::applyTo(QApplication& app) {
    app.setStyle(QStyleFactory::create("Fusion"));
    app.setPalette(paletteForTheme(theme()));

    QFile f(qssResourceForTheme(theme()));
    if (f.open(QIODevice::ReadOnly | QIODevice::Text)) {
        app.setStyleSheet(QString::fromUtf8(f.readAll()));
    }
}

QString ThemeManager::themeDisplayName(ThemeManager::Theme theme) {
    return theme == ThemeManager::Theme::Dark ? QStringLiteral("Тёмная") : QStringLiteral("Светлая");
}

QString ThemeManager::settingsKey() {
    return QStringLiteral("ui/theme");
}

ThemeManager::Theme ThemeManager::themeFromString(const QString& value) {
    const auto v = value.trimmed().toLower();
    if (v == QStringLiteral("light")) {
        return ThemeManager::Theme::Light;
    }
    return ThemeManager::Theme::Dark;
}

QString ThemeManager::themeToString(ThemeManager::Theme theme) {
    return theme == ThemeManager::Theme::Dark ? QStringLiteral("dark") : QStringLiteral("light");
}
