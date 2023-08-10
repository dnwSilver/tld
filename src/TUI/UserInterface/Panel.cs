using TUI.Settings;


namespace TUI.UserInterface;

public static class Panel
{
    private const int BorderWidth = 1;
    private const int ColumnWidth = 10;
    private const int TitleWidth = 35;
    private const int TagCount = 5;
    private const int TagWidth = 2;

    public static void RenderRows(SourceDto[] sources, int selectedRowNumber)
    {
        for (var index = 0; index < sources.Length; index++)
        {
            Console.SetCursorPosition(Theme.Padding,
                    6 + index + _marginTop + BorderWidth +
                    Theme.Padding);

            if (selectedRowNumber == index + 1)
            {
                // resultText = resultText.PastelBg("292928");
            }

            // Console.Write(resultText);
        }

        for (var index = 0; index < sources.Length; index++)
        {
            Console.SetCursorPosition(TitleWidth,
                    6 + index + _marginTop + BorderWidth + Theme.Padding);
            // var source = sources[index];
            // var package = DownloadPackage(source);
            // var resultText = package.Dependencies.React;
            // resultText = new string(' ', ColumnWidth - resultText.Width()) + resultText;
            // if (selectedRowNumber == index + 1)
            // {
            //     resultText = resultText.PastelBg("292928");
            // }
            //
            // Console.Write(resultText);
        }
        // for (var index = 0; index < sources.Length; index++)
        // {
        //     var loading = true;
        //     var braille = new[] { "⠿", "⠧", "⠏", "⠛", "⠹", "⠼", "⠶" };
        //     var braileNumber = 0;
        //     do
        //     {
        //         var resultText = braille[braileNumber];
        //         if (selectedRowNumber == index + 1)
        //         {
        //             resultText = resultText.PastelBg("292928");
        //         }
        //
        //         Console.SetCursorPosition(ColumnWidth + TagCountInLeftPanel * 2, index + 2);
        //         Console.Write(resultText);
        //         Thread.Sleep(100);
        //         if (braileNumber == braille.Length - 1)
        //         {
        //             braileNumber = 0;
        //             loading = false;
        //         }
        //         else
        //         {
        //             braileNumber++;
        //         }
        //     } while (loading);
        //     
        //     Console.SetCursorPosition(ColumnWidth + TagCountInLeftPanel * 2, index + 2);
        //     Console.Write(braille[0]);
        // }
    }


    private static int _marginTop;

    // private static Package DownloadPackage(Source source)
    // {
    //     if (Packages.TryGetValue(source.Repo, out var downloadPackage))
    //     {
    //         return downloadPackage;
    //     }
    //
    //     using HttpClient client = new();
    //     var json = client.GetStringAsync(source.Repo).GetAwaiter().GetResult();
    //     var package = JsonSerializer.Deserialize<Package>(json);
    //     Packages.Add(source.Repo, package);
    //     return package;
    // }
}