﻿using TUI.Pages;


Console.Clear();
Console.CursorVisible = false;

var welcomePage = WelcomePage.Instance;
welcomePage.Open();
Thread.Sleep(500);

IPage currentPage = DependenciesPage.Instance;
currentPage.Open();


ConsoleKeyInfo? key = null;

var waitCommand = true;
do
{
    switch (key?.Key)
    {
        case ConsoleKey.Q:
            waitCommand = false;
            break;
        case ConsoleKey.R:
            key = null;
            currentPage.Load();
            break;
        case ConsoleKey.D1:
            key = null;
            currentPage = DependenciesPage.Instance;
            Console.Clear();
            currentPage.Render();
            break;
        case ConsoleKey.D0:
            key = null;
            currentPage = WelcomePage.Instance;
            Console.Clear();
            currentPage.Render();
            break;
        default:
            key = Console.ReadKey(true);
            break;
    }
    
    // if (display.CommandLine.IsFocused)
    // {
    //     switch (key.Key)
    //     {
    //         case ConsoleKey.Escape:
    //             display.CommandLine.IsFocused = false;
    //             break;
    //         default:
    //             Console.Write(key.KeyChar);
    //
    //             break;
    //     }
    // }
    // else
    // {
    //     switch (key.KeyChar)
    //     {
    //         case ':':
    //             display.OpenCommandLine();
    //             break;
    //     }
    //
    //     switch (key.Key)
    //     {
    //         case ConsoleKey.DownArrow:
    //             display.Next();
    //             break;
    //         case ConsoleKey.UpArrow:
    //             display.Previous();
    //             break;
    //         case ConsoleKey.E:
    //             display.Toggle();
    //             break;
    //     }
    // }
} while (waitCommand);

Console.Clear();