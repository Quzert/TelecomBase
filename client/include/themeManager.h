#pragma once

#include <QString>

class QApplication;

class ThemeManager final {
public:
    enum class Theme { Light, Dark };

    static Theme theme();
    static void setTheme(Theme theme);

    static void applyTo(QApplication& app);

    static QString themeDisplayName(Theme theme);

private:
    static QString settingsKey();
    static Theme themeFromString(const QString& value);
    static QString themeToString(Theme theme);
};
