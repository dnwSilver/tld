using YamlDotNet.Serialization;


namespace TUI.Domain;

[YamlSerializable]
public class Source
{
    [YamlMember]
    public string[] Tags { get; set; }

    [YamlMember]
    public string Name { get; set; }

    [YamlMember]
    // [YamlMember(Alias = "project_id")]
    public int ProjectId { get; set; } = 0;

    [YamlMember]
    public string Repo { get; set; }
}